package envelopes

import (
	acservices "github.com/vonnwang/account/services"
	"github.com/vonnwang/resk/services"
	"github.com/segmentio/ksuid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	"strconv"
	"sync"
	"testing"
)

func TestRedEnvelopeService_Receive(t *testing.T) {
	//1. 准备几个红包资金账户，用于发红包和收红包
	accountService := acservices.GetAccountService()

	Convey("收红包测试用例", t, func() {
		accounts := make([]*acservices.AccountDTO, 0)
		size := 10
		for i := 0; i < size; i++ {
			account := acservices.AccountCreatedDTO{
				UserId:       ksuid.New().Next().String(),
				Username:     "测试用户" + strconv.Itoa(i+1),
				Amount:       "2000",
				AccountName:  "测试账户" + strconv.Itoa(i+1),
				AccountType:  int(acservices.EnvelopeAccountType),
				CurrencyCode: "CNY",
			}
			//账户创建
			acDto, err := accountService.CreateAccount(account)
			So(err, ShouldBeNil)
			So(acDto, ShouldNotBeNil)
			accounts = append(accounts, acDto)
		}
		acDto := accounts[0]
		So(len(accounts), ShouldEqual, size)
		//2. 使用其中一个用户发送一个红包
		re := services.GetRedEnvelopeService()
		//发送普通红包
		goods := services.RedEnvelopeSendingDTO{
			UserId:       acDto.UserId,
			Username:     acDto.Username,
			EnvelopeType: services.GeneralEnvelopeType,
			Amount:       decimal.NewFromFloat(1.88),
			Quantity:     size,
			Blessing:     services.DefaultBlessing,
		}
		at, err := re.SendOut(goods)
		So(err, ShouldBeNil)
		So(at, ShouldNotBeNil)
		So(at.Link, ShouldNotBeEmpty)
		So(at.RedEnvelopeGoodsDTO, ShouldNotBeNil)
		//验证每一个属性
		dto := at.RedEnvelopeGoodsDTO
		So(dto.Username, ShouldEqual, goods.Username)
		So(dto.UserId, ShouldEqual, goods.UserId)
		So(dto.Quantity, ShouldEqual, goods.Quantity)
		q := decimal.NewFromFloat(float64(dto.Quantity))
		So(dto.Amount.String(), ShouldEqual, goods.Amount.Mul(q).String())
		remainAmount := at.Amount
		//3. 使用发送红包数量的人收红包
		Convey("收普通红包", func() {
			for i, account := range accounts {
				rcv := services.RedEnvelopeReceiveDTO{
					EnvelopeNo:   at.EnvelopeNo,
					RecvUserId:   account.UserId,
					RecvUsername: account.Username,
					AccountNo:    account.AccountNo,
				}
				item, err := re.Receive(rcv)
				logrus.Info(i)
				logrus.Infof("%+v", item)
				So(err, ShouldBeNil)
				So(item, ShouldNotBeNil)
				So(item.Amount, ShouldEqual, at.AmountOne)
				remainAmount = remainAmount.Sub(at.AmountOne)
				So(item.RemainAmount.String(), ShouldEqual, remainAmount.String())

			}
		})

		//收碰运气红包，作为作业留给同学们来实现
		goods.EnvelopeType = services.LuckyEnvelopeType
		goods.Amount = decimal.NewFromFloat(18.8)
		at, err = re.SendOut(goods)
		So(err, ShouldBeNil)
		So(at, ShouldNotBeNil)
		So(at.Link, ShouldNotBeEmpty)
		So(at.RedEnvelopeGoodsDTO, ShouldNotBeNil)
		//验证每一个属性
		dto = at.RedEnvelopeGoodsDTO
		So(dto.Username, ShouldEqual, goods.Username)
		So(dto.UserId, ShouldEqual, goods.UserId)
		So(dto.Quantity, ShouldEqual, goods.Quantity)
		So(dto.Amount.String(), ShouldEqual, goods.Amount.String())
		remainAmount = at.Amount
		re = services.GetRedEnvelopeService()
		Convey("收碰运气红包", func() {
			So(len(accounts), ShouldEqual, size)
			total := decimal.NewFromFloat(0)
			for i, account := range accounts {
				if i > 10 {
					break
				}
				rcv := services.RedEnvelopeReceiveDTO{
					EnvelopeNo:   at.EnvelopeNo,
					RecvUserId:   account.UserId,
					RecvUsername: account.Username,
					AccountNo:    account.AccountNo,
				}
				item, err := re.Receive(rcv)
				if item != nil {
					total = total.Add(item.Amount)
				}

				logrus.Info(i+1, " ", total.String(), " ", item.Amount.String())

				So(err, ShouldBeNil)
				So(item, ShouldNotBeNil)
				remainAmount = remainAmount.Sub(item.Amount)
				So(item.RemainAmount.String(), ShouldEqual, remainAmount.String())

			}
			So(total.String(), ShouldEqual, goods.Amount.String())
		})

	})

}

func TestRedEnvelopeService_Receive_Failure(t *testing.T) {
	//1. 准备几个红包资金账户，用于发红包和收红包
	accountService := acservices.GetAccountService()

	Convey("收红包测试用例", t, func() {
		accounts := make([]*acservices.AccountDTO, 0)
		size := 5
		for i := 0; i < size; i++ {
			account := acservices.AccountCreatedDTO{
				UserId:       ksuid.New().Next().String(),
				Username:     "测试用户" + strconv.Itoa(i+1),
				Amount:       "100",
				AccountName:  "测试账户" + strconv.Itoa(i+1),
				AccountType:  int(acservices.EnvelopeAccountType),
				CurrencyCode: "CNY",
			}
			//账户创建
			acDto, err := accountService.CreateAccount(account)
			So(err, ShouldBeNil)
			So(acDto, ShouldNotBeNil)
			accounts = append(accounts, acDto)
		}
		//2. 使用其中一个用户发送一个红包
		acDto := accounts[0]
		So(len(accounts), ShouldEqual, size)
		re := services.GetRedEnvelopeService()
		//发送普通红包
		goods := services.RedEnvelopeSendingDTO{
			UserId:       acDto.UserId,
			Username:     acDto.Username,
			EnvelopeType: services.LuckyEnvelopeType,
			Amount:       decimal.NewFromFloat(10),
			Quantity:     3,
			Blessing:     services.DefaultBlessing,
		}
		at, err := re.SendOut(goods)
		So(err, ShouldBeNil)
		So(at, ShouldNotBeNil)
		So(at.Link, ShouldNotBeEmpty)
		So(at.RedEnvelopeGoodsDTO, ShouldNotBeNil)
		//验证每一个属性
		dto := at.RedEnvelopeGoodsDTO
		So(dto.Username, ShouldEqual, goods.Username)
		So(dto.UserId, ShouldEqual, goods.UserId)
		So(dto.Quantity, ShouldEqual, goods.Quantity)
		So(dto.Amount.String(), ShouldEqual, goods.Amount.String())
		//
		re = services.GetRedEnvelopeService()
		Convey("收碰运气红包", func() {
			So(len(accounts), ShouldEqual, size)
			total := decimal.NewFromFloat(0)
			remainAmount := goods.Amount
			sendingAmount := decimal.NewFromFloat(0)

			for i, account := range accounts {
				rcv := services.RedEnvelopeReceiveDTO{
					EnvelopeNo:   at.EnvelopeNo,
					RecvUserId:   account.UserId,
					RecvUsername: account.Username,
					AccountNo:    account.AccountNo,
				}
				if i <= 2 {
					item, err := re.Receive(rcv)
					if item != nil {
						total = total.Add(item.Amount)
					}
					logrus.Info(i+1, " ", total.String(), " ", item.Amount.String())
					So(err, ShouldBeNil)
					So(item, ShouldNotBeNil)
					remainAmount = remainAmount.Sub(item.Amount)
					So(item.RemainAmount.String(), ShouldEqual, remainAmount.String())
					a := accountService.GetEnvelopeAccountByUserId(rcv.RecvUserId)
					So(a, ShouldNotBeNil)
					if item.RecvUserId == goods.UserId {
						b := decimal.NewFromFloat(100)
						b = b.Sub(decimal.NewFromFloat(10))
						b = b.Add(item.Amount)
						So(a.Balance.String(), ShouldEqual, b.String())
						sendingAmount = item.Amount
					} else {
						So(a.Balance.String(), ShouldEqual, item.Amount.Add(decimal.NewFromFloat(100)).String())
					}

				} else {
					item, err := re.Receive(rcv)
					So(err, ShouldNotBeNil)
					So(item, ShouldBeNil)
				}

			}
			So(total.String(), ShouldEqual, goods.Amount.String())

			order := re.Get(at.EnvelopeNo)
			So(order, ShouldNotBeNil)
			So(order.RemainAmount.String(), ShouldEqual, "0")
			So(order.RemainQuantity, ShouldEqual, 0)
			a := accountService.GetEnvelopeAccountByUserId(order.UserId)
			So(a, ShouldNotBeNil)
			So(a.Balance.String(), ShouldEqual, sendingAmount.Add(decimal.NewFromFloat(90)).String())
		})

	})

}

func TestRedEnvelopeService_Receive_C(t *testing.T) {
	//1. 准备几个红包资金账户，用于发红包和收红包
	accountService := acservices.GetAccountService()

	Convey("收红包测试用例", t, func() {
		accounts := make([]*acservices.AccountDTO, 0)
		size := 100
		for i := 0; i < size; i++ {
			account := acservices.AccountCreatedDTO{
				UserId:       ksuid.New().Next().String(),
				Username:     "测试用户" + strconv.Itoa(i+1),
				Amount:       "2000",
				AccountName:  "测试账户" + strconv.Itoa(i+1),
				AccountType:  int(acservices.EnvelopeAccountType),
				CurrencyCode: "CNY",
			}
			//账户创建
			acDto, err := accountService.CreateAccount(account)
			So(err, ShouldBeNil)
			So(acDto, ShouldNotBeNil)
			accounts = append(accounts, acDto)
		}
		acDto := accounts[0]
		So(len(accounts), ShouldEqual, size)
		//2. 使用其中一个用户发送一个红包
		re := services.GetRedEnvelopeService()
		//发送普通红包
		goods := services.RedEnvelopeSendingDTO{
			UserId:       acDto.UserId,
			Username:     acDto.Username,
			EnvelopeType: services.LuckyEnvelopeType,
			Amount:       decimal.NewFromFloat(20),
			Quantity:     10,
			Blessing:     services.DefaultBlessing,
		}
		at, err := re.SendOut(goods)
		So(err, ShouldBeNil)
		So(at, ShouldNotBeNil)
		So(at.Link, ShouldNotBeEmpty)
		So(at.RedEnvelopeGoodsDTO, ShouldNotBeNil)
		//验证每一个属性
		dto := at.RedEnvelopeGoodsDTO
		So(dto.Username, ShouldEqual, goods.Username)
		So(dto.UserId, ShouldEqual, goods.UserId)
		So(dto.Quantity, ShouldEqual, goods.Quantity)
		So(dto.Amount.String(), ShouldEqual, goods.Amount.String())
		at, err = re.SendOut(goods)
		So(err, ShouldBeNil)
		So(at, ShouldNotBeNil)
		So(at.Link, ShouldNotBeEmpty)
		So(at.RedEnvelopeGoodsDTO, ShouldNotBeNil)
		//验证每一个属性
		dto = at.RedEnvelopeGoodsDTO
		So(dto.Username, ShouldEqual, goods.Username)
		So(dto.UserId, ShouldEqual, goods.UserId)
		So(dto.Quantity, ShouldEqual, goods.Quantity)
		So(dto.Amount.String(), ShouldEqual, goods.Amount.String())
		re = services.GetRedEnvelopeService()
		Convey("收碰运气红包", func() {
			So(len(accounts), ShouldEqual, size)
			var wg sync.WaitGroup
			for _, account := range accounts {
				wg.Add(1)
				go func() {
					rcv := services.RedEnvelopeReceiveDTO{
						EnvelopeNo:   at.EnvelopeNo,
						RecvUserId:   account.UserId,
						RecvUsername: account.Username,
						AccountNo:    account.AccountNo,
					}
					re.Receive(rcv)
					wg.Done()
				}()

			}
			wg.Wait()
			items := re.ListItems(at.EnvelopeNo)
			So(len(items), ShouldEqual, 10)
			total := decimal.NewFromFloat(0)
			for _, item := range items {
				total = total.Add(item.Amount)
			}
			So(total.String(), ShouldEqual, at.Amount.String())
			//这里就测试出来一个并发下面的临界问题：
			//当库存还大于1的时候，假如剩余的金额为2元
			// 那么在计算红包序列的时候是不会取最后剩余的所有金额的
			//仍然是以数量大于1的情况下计算的，
			// 比如有2个用户同时抢，分别计算得到0.8元和1.1元
			//这时1.1元的用户获得红包，同时还剩余0.9元，并且库存是1
			// 这个时候0.8元的用户也去抢，符合乐观锁的逻辑，也抢到0.8元，
			// 然后库存剩余为0，但剩余金额为0.1元，之后由于库存为0了，
			// 剩余金额就不能被抢走了，那么这个问题呢通过乐观锁是不能解决的，
			// 我这里把这个问题留给同学们思考一下，查阅一下资料看如何来解决？
			// 在问答区我们进行讨论，看看有多少方法可以解决这个问题？
			//
		})

	})

}
