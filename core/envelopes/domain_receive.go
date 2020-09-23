package envelopes

import (
	"context"
	"database/sql"
	"errors"
	"github.com/vonnwang/account/core/accounts"
	acservices "github.com/vonnwang/account/services"
	"github.com/vonnwang/infra/algo"
	"github.com/vonnwang/infra/base"
	"github.com/vonnwang/resk/services"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/tietang/dbx"
)

var multiple = decimal.NewFromFloat(100.0)

//收红包业务逻辑代码
func (d *goodsDomain) Receive(
	ctx context.Context,
	dto services.RedEnvelopeReceiveDTO) (item *services.RedEnvelopeItemDTO, err error) {
	//1.创建收红包的订单明细 preCreatItem
	d.preCreateItem(dto)
	//2.查询出当前红包的剩余数量和剩余金额信息
	goods := d.Get(dto.EnvelopeNo)
	//3. 效验剩余红包和剩余金额：
	//- 如果没有剩余，直接返回无可用红包金额
	if goods.RemainQuantity <= 0 || goods.RemainAmount.Cmp(decimal.NewFromFloat(0)) <= 0 {
		log.Errorf("没有足够的红包和金额了: %+v", goods)
		return nil, errors.New("没有足够的红包和金额了")
	}

	//4. 使用红包算法计算红包金额
	nextAmount := d.nextAmount(goods)
	err = base.Tx(func(runner *dbx.TxRunner) error {
		//5. 使用乐观锁更新语句，尝试更新剩余数量和剩余金额：
		dao := RedEnvelopeGoodsDao{runner: runner}
		rows, err := dao.UpdateBalance(goods.EnvelopeNo, nextAmount)

		// - 如果更新失败，也就是返回0，表示无可用红包数量和金额，抢红包失败
		if rows <= 0 || err != nil {
			return errors.New("没有足够的红包和金额了")
		}
		// - 如果更新成功，也就是返回1，表示抢到红包
		//6. 保存订单明细数据
		d.item.Quantity = 1
		d.item.PayStatus = int(services.Paying)
		d.item.AccountNo = dto.AccountNo
		d.item.RemainAmount = goods.RemainAmount.Sub(nextAmount)
		d.item.Amount = nextAmount
		desc := goods.Username.String + "的" + services.EnvelopeTypes[services.EnvelopeType(goods.EnvelopeType)]
		d.item.Desc = desc
		txCtx := base.WithValueContext(ctx, runner)
		_, err = d.item.Save(txCtx)
		//insert into `red_envelope_item`(`item_no`,`envelope_no`,`amount`,`quantity`,`remain_amount`,`account_no`,`recv_username`,`recv_user_id`,`pay_status`)
		// values(?,?,?,?,?,?,?,?,?)
		// 1LlS0fIHk4UxdnqC8wNGww5V7cJ
		//1LlU0eDA6WNt16ZbLKp3RwpFWcp
		if err != nil {
			log.Error(err)
			return err
		}
		//7. 将抢到的红包金额从系统红包中间账户转入当前用户的资金账户
		// transfer
		status, err := d.transfer(txCtx, dto)
		if status == acservices.TransferedStatusSuccess {
			return nil
		} else {
			return err
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return d.item.ToDTO(), err
}

func (d *goodsDomain) transfer(
	ctx context.Context,
	dto services.RedEnvelopeReceiveDTO) (status acservices.TransferedStatus, err error) {
	systemAccount := base.GetSystemAccount()
	body := acservices.TradeParticipator{
		AccountNo: systemAccount.AccountNo,
		UserId:    systemAccount.UserId,
		Username:  systemAccount.Username,
	}
	target := acservices.TradeParticipator{
		AccountNo: dto.AccountNo,
		UserId:    dto.RecvUserId,
		Username:  dto.RecvUsername,
	}

	adomain := accounts.NewAccountDomain()
	//从系统红包资金账户扣减
	transfer := acservices.AccountTransferDTO{
		TradeBody:   body,
		TradeTarget: target,
		TradeNo:     dto.EnvelopeNo,
		Amount:      d.item.Amount,
		ChangeType:  acservices.EnvelopeOutgoing,
		ChangeFlag:  acservices.FlagTransferOut,
		Decs:        "红包扣减：" + dto.EnvelopeNo,
	}
	status, err = adomain.TransferWithContextTx(ctx, transfer)
	if err != nil || status != acservices.TransferedStatusSuccess {
		return status, err
	}
	//从系统红包资金账户转入当前用户
	transfer = acservices.AccountTransferDTO{
		TradeBody:   target,
		TradeTarget: body,
		TradeNo:     dto.EnvelopeNo,
		Amount:      d.item.Amount,
		ChangeType:  acservices.EnvelopeIncoming,
		ChangeFlag:  acservices.FlagTransferIn,
		Decs:        "红包收入：" + dto.EnvelopeNo,
	}
	return adomain.TransferWithContextTx(ctx, transfer)
}

//预创建收红包订单明细
func (d *goodsDomain) preCreateItem(dto services.RedEnvelopeReceiveDTO) {
	d.item.AccountNo = dto.AccountNo
	d.item.EnvelopeNo = dto.EnvelopeNo
	d.item.RecvUsername = sql.NullString{String: dto.RecvUsername, Valid: true}
	d.item.RecvUserId = dto.RecvUserId
	d.item.createItemNo()
}

//计算红包金额
func (d *goodsDomain) nextAmount(goods *RedEnvelopeGoods) (amount decimal.Decimal) {
	if goods.RemainQuantity == 1 {
		return goods.RemainAmount
	}
	if goods.EnvelopeType == services.GeneralEnvelopeType {
		return goods.AmountOne
	} else if goods.EnvelopeType == services.LuckyEnvelopeType {
		cent := goods.RemainAmount.Mul(multiple).IntPart()
		next := algo.DoubleAverage(int64(goods.RemainQuantity), cent)
		amount = decimal.NewFromFloat(float64(next)).Div(multiple)
	} else {
		log.Error("不支持的红包类型")
	}
	return amount
}
