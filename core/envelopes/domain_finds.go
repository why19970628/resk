package envelopes

import (
	"github.com/vonnwang/infra/base"
	"github.com/tietang/dbx"
)

func (d *goodsDomain) Find(po *RedEnvelopeGoods, offset, limit int) (regs []RedEnvelopeGoods) {
	base.Tx(func(runner *dbx.TxRunner) error {
		dao := RedEnvelopeGoodsDao{runner: runner}
		regs = dao.Find(po, offset, limit)
		return nil
	})
	return regs
}

func (d *goodsDomain) FindByUser(userId string, offset, limit int) (regs []RedEnvelopeGoods) {
	base.Tx(func(runner *dbx.TxRunner) error {
		dao := RedEnvelopeGoodsDao{runner: runner}
		regs = dao.FindByUser(userId, offset, limit)
		return nil
	})
	return regs
}
func (d *goodsDomain) GetOne(envelopeNo string) (po *RedEnvelopeGoods) {
	base.Tx(func(runner *dbx.TxRunner) error {
		dao := RedEnvelopeGoodsDao{runner: runner}
		po = dao.GetOne(envelopeNo)
		return nil
	})
	return po
}

func (d *goodsDomain) ListReceivable(offset, limit int) (regs []RedEnvelopeGoods) {
	base.Tx(func(runner *dbx.TxRunner) error {
		dao := RedEnvelopeGoodsDao{runner: runner}
		regs = dao.ListReceivable(offset, limit)
		return nil
	})
	return regs
}
func (d *goodsDomain) ListReceived(userId string, offset, limit int) (regs []*RedEnvelopeItem) {
	base.Tx(func(runner *dbx.TxRunner) error {
		dao := RedEnvelopeItemDao{runner: runner}
		regs = dao.ListReceivedItems(userId, offset, limit)
		return nil
	})
	return regs
}
