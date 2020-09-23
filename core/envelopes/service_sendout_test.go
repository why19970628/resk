package envelopes

import (
	acservices "github.com/vonnwang/account/services"
	"github.com/vonnwang/resk/services"
	_ "github.com/vonnwang/resk/testx"
	"github.com/segmentio/ksuid"
	"github.com/shopspring/decimal"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestRedEnvelopeService_SendOut(t *testing.T) {
	//发红包人的红包资金账户
	ac := acservices.GetAccountService()
	account := acservices.AccountCreatedDTO{
		UserId:       ksuid.New().Next().String(),
		Username:     "测试用户",
		Amount:       "10000",
		AccountName:  "测试账户",
		AccountType:  int(acservices.EnvelopeAccountType),
		CurrencyCode: "CNY",
	}
	re := services.GetRedEnvelopeService()
	Convey("准备资金账户", t, func() {
		//准备资金账户
		acDTO, err := ac.CreateAccount(account)
		So(err, ShouldBeNil)
		So(acDTO, ShouldNotBeNil)
	})
	Convey("发红红包	", t, func() {
		Convey("发普通红包", func() {
			goods := services.RedEnvelopeSendingDTO{
				UserId:       account.UserId,
				Username:     account.Username,
				EnvelopeType: services.GeneralEnvelopeType,
				Amount:       decimal.NewFromFloat(8.88),
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
			q := decimal.NewFromFloat(float64(dto.Quantity))
			So(dto.Amount.String(), ShouldEqual, goods.Amount.Mul(q).String())
			//同学可以想一下，还需要验证哪些字段
		})
		Convey("发碰运气红包", func() {
			goods := services.RedEnvelopeSendingDTO{
				UserId:       account.UserId,
				Username:     account.Username,
				EnvelopeType: services.LuckyEnvelopeType,
				Amount:       decimal.NewFromFloat(88.8),
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
			//同学可以想一下，还需要验证哪些字段
		})
	})

}

func TestRedEnvelopeService_SendOut_Failure(t *testing.T) {
	//发红包人的红包资金账户
	ac := acservices.GetAccountService()
	account := acservices.AccountCreatedDTO{
		UserId:       ksuid.New().Next().String(),
		Username:     "测试用户A",
		Amount:       "10",
		AccountName:  "测试账户A",
		AccountType:  int(acservices.EnvelopeAccountType),
		CurrencyCode: "CNY",
	}
	re := services.GetRedEnvelopeService()
	Convey("准备资金账户", t, func() {
		//准备资金账户
		acDTO, err := ac.CreateAccount(account)
		So(err, ShouldBeNil)
		So(acDTO, ShouldNotBeNil)
	})
	Convey("发红红包	", t, func() {
		Convey("发碰运气红包", func() {
			goods := services.RedEnvelopeSendingDTO{
				UserId:       account.UserId,
				Username:     account.Username,
				EnvelopeType: services.LuckyEnvelopeType,
				Amount:       decimal.NewFromFloat(11),
				Quantity:     10,
				Blessing:     services.DefaultBlessing,
			}
			at, err := re.SendOut(goods)
			So(err, ShouldNotBeNil)
			So(at, ShouldBeNil)
			a := ac.GetEnvelopeAccountByUserId(account.UserId)
			So(a, ShouldNotBeNil)
			So(a.Balance.String(), ShouldEqual, account.Amount)

		})

	})

}
