package envelopes

import (
	"context"
	"github.com/vonnwang/account/core/accounts"
	acservices "github.com/vonnwang/account/services"
	"github.com/vonnwang/infra/base"
	"github.com/vonnwang/resk/services"
	"github.com/tietang/dbx"
	"path"
)

//发红包业务领域代码
func (d *goodsDomain) SendOut(
	goods services.RedEnvelopeGoodsDTO) (activity *services.RedEnvelopeActivity, err error) {
	//创建红包商品
	d.Create(goods)
	//创建活动
	activity = new(services.RedEnvelopeActivity)
	//红包链接，格式：http://域名/v1/envelope/{id}/link/
	link := base.GetEnvelopeActivityLink()
	domain := base.GetEnvelopeDomain()
	activity.Link = path.Join(domain, link, d.EnvelopeNo)
	accountDomain := accounts.NewAccountDomain()

	err = base.Tx(func(runner *dbx.TxRunner) (err error) {
		ctx := base.WithValueContext(context.Background(), runner)

		//事务逻辑问题：
		//保存红包商品和红包金额的支付必须要保证全部成功或者全部失败

		//保存红包商品
		id, err := d.Save(ctx)
		if id <= 0 || err != nil {
			return err
		}
		//红包金额支付
		//1. 需要红包中间商的红包资金账户，定义在配置文件中，事先初始化到资金账户表中
		//2. 从红包发送人的资金账户中扣减红包金额 ，把红包金额从红包发送人的资金账户里扣除
		body := acservices.TradeParticipator{
			AccountNo: goods.AccountNo,
			UserId:    goods.UserId,
			Username:  goods.Username,
		}
		systemAccount := base.GetSystemAccount()
		target := acservices.TradeParticipator{
			AccountNo: systemAccount.AccountNo,
			Username:  systemAccount.Username,
			UserId:    systemAccount.UserId,
		}

		transfer := acservices.AccountTransferDTO{
			TradeBody:   body,
			TradeTarget: target,
			TradeNo:     d.EnvelopeNo,
			Amount:      d.Amount,
			ChangeType:  acservices.EnvelopeOutgoing,
			ChangeFlag:  acservices.FlagTransferOut,
			Decs:        "红包金额支付",
		}
		status, err := accountDomain.TransferWithContextTx(ctx, transfer)
		if status != acservices.TransferedStatusSuccess {
			return err
		}
		//3. 将扣减的红包总金额转入红包中间商的红包资金账户
		//入账
		transfer = acservices.AccountTransferDTO{
			TradeBody:   target,
			TradeTarget: body,
			TradeNo:     d.EnvelopeNo,
			Amount:      d.Amount,
			ChangeType:  acservices.EnvelopeIncoming,
			ChangeFlag:  acservices.FlagTransferIn,
			Decs:        "红包金额转入",
		}
		status, err = accountDomain.TransferWithContextTx(ctx, transfer)
		if status == acservices.TransferedStatusSuccess {
			return nil
		} else {
			return err
		}
		_, err = d.UpdatePayStatus(d.EnvelopeNo, services.Payed)
		return err
	})
	if err != nil {
		return nil, err
	}

	//扣减金额没有问题，返回活动

	activity.RedEnvelopeGoodsDTO = *d.RedEnvelopeGoods.ToDTO()

	return activity, err
}
