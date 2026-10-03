package dao

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"semi-mes/server/internal/model"
)

// CarrierView is a carrier row plus the lot currently bound to it.
type CarrierView struct {
	model.WipCarrier
	LotNo string `json:"lotNo" gorm:"column:lot_no"`
	LotID uint64 `json:"lotId" gorm:"column:lot_id"`
}

// SlotCell is one position in a carrier.
type SlotCell struct {
	Slot    int    `json:"slot"`
	WaferID uint64 `json:"waferId"`
	WaferNo string `json:"waferNo"`
	Status  string `json:"status"`
	LotNo   string `json:"lotNo"`
}

func ListCarriers(db *gorm.DB, page, limit int, sort string, columns []Column) ([]CarrierView, int64, error) {
	q := db.Table("wip_carrier c").
		Select("c.*, l.lot_no, b.lot_id").
		Joins("LEFT JOIN wip_carrier_bind b ON b.carrier_id = c.id AND b.status = ?", model.BindBound).
		Joins("LEFT JOIN wip_lot l ON l.id = b.lot_id AND l.deleted_at IS NULL").
		Where("c.deleted_at IS NULL")
	q = applyColumns(q, columns, map[string]string{
		"carrier_no": "c.carrier_no", "carrier_type": "c.carrier_type", "status": "c.status",
		"location": "c.location", "lot_no": "l.lot_no", "created_at": "c.created_at",
	})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []CarrierView
	err := q.Order(orderClause(sort, map[string]string{
		"id": "c.id", "carrier_no": "c.carrier_no", "status": "c.status", "created_at": "c.created_at",
	}, "c.id DESC")).Offset(page * limit).Limit(limit).Scan(&rows).Error
	if rows == nil {
		rows = []CarrierView{}
	}
	return rows, total, err
}

func SaveCarrier(db *gorm.DB, row *model.WipCarrier) error {
	row.CarrierNo = strings.TrimSpace(row.CarrierNo)
	row.CarrierType = strings.TrimSpace(row.CarrierType)
	row.Status = strings.TrimSpace(row.Status)
	row.Location = strings.TrimSpace(row.Location)
	if row.CarrierType == "" {
		row.CarrierType = "FOUP"
	}
	if row.Capacity <= 0 {
		row.Capacity = 25
	}
	if row.Capacity > 25 {
		row.Capacity = 25
	}
	if row.Status == "" {
		row.Status = model.CarrierEmpty
	}
	switch row.Status {
	case model.CarrierEmpty, model.CarrierInUse, model.CarrierCleaning, model.CarrierDown:
	default:
		return ErrWipState
	}
	if row.CleanCount < 0 || row.CleanLimit < 0 {
		return ErrWipQty
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if row.CarrierNo == "" {
			no, err := Allocate(tx, model.RuleCarrier, time.Now())
			if err != nil {
				return err
			}
			row.CarrierNo = no
		}
		var existing model.WipCarrier
		err := tx.Where("carrier_no = ?", row.CarrierNo).First(&existing).Error
		if err == nil && existing.ID != row.ID {
			return ErrWipState
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if row.ID == 0 {
			return tx.Create(row).Error
		}
		return tx.Model(&model.WipCarrier{}).Where("id = ?", row.ID).Updates(map[string]any{
			"carrier_no": row.CarrierNo, "carrier_type": row.CarrierType, "capacity": row.Capacity,
			"status": row.Status, "location": row.Location, "clean_count": row.CleanCount,
			"clean_limit": row.CleanLimit, "note": row.Note,
		}).Error
	})
}

func BindCarrier(db *gorm.DB, carrierID, lotID uint64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var carrier model.WipCarrier
		if err := tx.First(&carrier, carrierID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrWipNotFound
			}
			return err
		}
		if carrier.Status == model.CarrierCleaning || carrier.Status == model.CarrierDown {
			return ErrWipState
		}
		var used int64
		if err := tx.Model(&model.WipCarrierBind{}).Where("carrier_id = ? AND status = ?", carrier.ID, model.BindBound).Count(&used).Error; err != nil {
			return err
		}
		if used > 0 {
			return ErrWipState
		}
		if err := tx.Model(&model.WipCarrierBind{}).Where("lot_id = ? AND status = ?", lotID, model.BindBound).Count(&used).Error; err != nil {
			return err
		}
		if used > 0 {
			return ErrWipState
		}
		lot, err := GetLot(tx, lotID)
		if err != nil {
			return err
		}
		if lot.Status != model.LotWaiting && lot.Status != model.LotHold {
			return ErrWipState
		}
		wafers, err := activeWafers(tx, lot.ID)
		if err != nil {
			return err
		}
		if len(wafers) > carrier.Capacity {
			return ErrWipQty
		}
		if err := tx.Create(&model.WipCarrierBind{CarrierID: carrier.ID, LotID: lot.ID, Status: model.BindBound}).Error; err != nil {
			return err
		}
		if err := tx.Model(&carrier).Update("status", model.CarrierInUse).Error; err != nil {
			return err
		}
		return placeWafers(tx, wafers, lot.ID, carrier.ID, "bind")
	})
}

func UnbindCarrier(db *gorm.DB, carrierID uint64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var bind model.WipCarrierBind
		err := tx.Where("carrier_id = ? AND status = ?", carrierID, model.BindBound).First(&bind).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrWipState
		}
		if err != nil {
			return err
		}
		if err := tx.Model(&bind).Update("status", model.BindUnbound).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.WipCarrier{}).Where("id = ?", carrierID).Update("status", model.CarrierEmpty).Error; err != nil {
			return err
		}
		wafers, err := activeWafers(tx, bind.LotID)
		if err != nil {
			return err
		}
		return placeWafers(tx, wafers, bind.LotID, 0, "unbind")
	})
}

func SlotMap(db *gorm.DB, carrierID uint64) (*model.WipCarrier, []SlotCell, string, error) {
	var carrier model.WipCarrier
	if err := db.First(&carrier, carrierID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, "", ErrWipNotFound
		}
		return nil, nil, "", err
	}
	slots := make([]SlotCell, carrier.Capacity)
	for i := range slots {
		slots[i].Slot = i + 1
	}
	var lotNo string
	var bind model.WipCarrierBind
	err := db.Where("carrier_id = ? AND status = ?", carrier.ID, model.BindBound).First(&bind).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, "", err
	}
	if err == nil {
		var lot model.WipLot
		if err := db.First(&lot, bind.LotID).Error; err == nil {
			lotNo = lot.LotNo
		}
	}
	var wafers []model.WipWafer
	if err := db.Where("carrier_id = ? AND status = ?", carrier.ID, model.WaferActive).Find(&wafers).Error; err != nil {
		return nil, nil, "", err
	}
	for _, wafer := range wafers {
		if wafer.Slot < 1 || wafer.Slot > len(slots) {
			continue
		}
		cell := &slots[wafer.Slot-1]
		cell.WaferID = wafer.ID
		cell.WaferNo = wafer.WaferNo
		cell.Status = wafer.Status
		cell.LotNo = lotNo
	}
	return &carrier, slots, lotNo, nil
}

func ListLotWafers(db *gorm.DB, lotID uint64) ([]model.WipWafer, error) {
	var rows []model.WipWafer
	err := db.Where("lot_id = ?", lotID).Order("slot ASC, id ASC").Find(&rows).Error
	if rows == nil {
		rows = []model.WipWafer{}
	}
	return rows, err
}

func ScrapLotWafers(db *gorm.DB, lotID uint64, ids []uint64, reasonCode string) (*model.WipLot, error) {
	if len(ids) == 0 || !AcceptCode(db, "scrap", reasonCode, ScrapReasons) {
		return nil, ErrWipQty
	}
	var updated *model.WipLot
	err := db.Transaction(func(tx *gorm.DB) error {
		lot, err := GetLot(tx, lotID)
		if err != nil {
			return err
		}
		if lot.Status != model.LotWaiting && lot.Status != model.LotHold {
			return ErrWipState
		}
		if err := scrapWaferIDs(tx, lot.ID, ids, reasonCode, 0); err != nil {
			return err
		}
		lot.Quantity -= len(ids)
		if lot.Quantity < 0 {
			return ErrWipQty
		}
		if lot.Quantity == 0 {
			lot.Status = model.LotScrapped
			if err := releaseLotCarrier(tx, lot.ID); err != nil {
				return err
			}
		}
		if err := tx.Model(lot).Updates(map[string]any{"quantity": lot.Quantity, "status": lot.Status}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.WipLotHistory{
			LotID: lot.ID, EventType: model.EventScrap, FromNodeKey: lot.CurrentNodeKey, ToNodeKey: lot.CurrentNodeKey,
			ReasonCode: reasonCode, Quantity: len(ids),
		}).Error; err != nil {
			return err
		}
		updated = lot
		return nil
	})
	return updated, err
}

func createLotWafers(tx *gorm.DB, lot *model.WipLot, now time.Time) error {
	for i := 1; i <= lot.Quantity; i++ {
		no, err := Allocate(tx, model.RuleWafer, now)
		if err != nil {
			return err
		}
		wafer := model.WipWafer{WaferNo: no, LotID: lot.ID, Slot: i, Status: model.WaferActive, CreatedAt: &now}
		if err := tx.Create(&wafer).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.WipWaferHistory{
			WaferID: wafer.ID, LotID: lot.ID, EventType: model.EventStart, ToSlot: i, CreatedAt: &now,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func moveSplitWafers(tx *gorm.DB, parent *model.WipLot, child *model.WipLot, qty int, ids []uint64) error {
	total, err := countActive(tx, parent.ID)
	if err != nil || total == 0 {
		return err
	}
	picked, err := pickWafers(tx, parent.ID, qty, ids)
	if err != nil {
		return err
	}
	now := time.Now()
	for i, wafer := range picked {
		slot := i + 1
		if err := tx.Model(&model.WipWafer{}).Where("id = ?", wafer.ID).Updates(map[string]any{
			"lot_id": child.ID, "carrier_id": 0, "slot": slot,
		}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.WipWaferHistory{
			WaferID: wafer.ID, LotID: child.ID, EventType: model.EventSplit, FromSlot: wafer.Slot, ToSlot: slot, CreatedAt: &now,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func finishSplitSlots(tx *gorm.DB, parentID uint64) error {
	total, err := countActive(tx, parentID)
	if err != nil || total == 0 {
		return err
	}
	return compactSlots(tx, parentID)
}

func releaseLotCarrier(tx *gorm.DB, lotID uint64) error {
	var bind model.WipCarrierBind
	err := tx.Where("lot_id = ? AND status = ?", lotID, model.BindBound).First(&bind).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := tx.Model(&bind).Update("status", model.BindUnbound).Error; err != nil {
		return err
	}
	return tx.Model(&model.WipCarrier{}).Where("id = ?", bind.CarrierID).Update("status", model.CarrierEmpty).Error
}

func moveMergeWafers(tx *gorm.DB, sourceID, targetID uint64) error {
	if err := releaseLotCarrier(tx, sourceID); err != nil {
		return err
	}
	total, err := countActive(tx, sourceID)
	if err != nil || total == 0 {
		return err
	}
	wafers, err := activeWafers(tx, sourceID)
	if err != nil {
		return err
	}
	now := time.Now()
	for _, wafer := range wafers {
		if err := tx.Model(&model.WipWafer{}).Where("id = ?", wafer.ID).Updates(map[string]any{
			"lot_id": targetID, "carrier_id": 0, "slot": 0,
		}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.WipWaferHistory{
			WaferID: wafer.ID, LotID: targetID, EventType: model.EventMerge, FromSlot: wafer.Slot, CreatedAt: &now,
		}).Error; err != nil {
			return err
		}
	}
	return compactSlots(tx, targetID)
}

func validateBoundCarrier(tx *gorm.DB, lotID uint64, carrierNo string) error {
	carrierNo = strings.TrimSpace(carrierNo)
	var bind model.WipCarrierBind
	err := tx.Where("lot_id = ? AND status = ?", lotID, model.BindBound).First(&bind).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if carrierNo != "" {
			return ErrWipTrack
		}
		return nil
	}
	if err != nil {
		return err
	}
	var carrier model.WipCarrier
	if err := tx.First(&carrier, bind.CarrierID).Error; err != nil {
		return err
	}
	if carrierNo != "" && !strings.EqualFold(carrier.CarrierNo, carrierNo) {
		return ErrWipTrack
	}
	if carrier.Status != model.CarrierInUse {
		return ErrWipTrack
	}
	if carrier.CleanLimit > 0 && carrier.CleanCount >= carrier.CleanLimit {
		return ErrWipTrack
	}
	n, err := countActive(tx, lotID)
	if err != nil {
		return err
	}
	if n > int64(carrier.Capacity) {
		return ErrWipTrack
	}
	return nil
}

func boundCarrierNo(db *gorm.DB, lotID uint64) string {
	var no string
	_ = db.Table("wip_carrier c").
		Joins("JOIN wip_carrier_bind b ON b.carrier_id = c.id AND b.status = ? AND b.lot_id = ?", model.BindBound, lotID).
		Where("c.deleted_at IS NULL").
		Limit(1).
		Pluck("c.carrier_no", &no).Error
	return no
}

func noteActiveWafers(tx *gorm.DB, lotID, moveID uint64, event, node string) error {
	wafers, err := activeWafers(tx, lotID)
	if err != nil || len(wafers) == 0 {
		return err
	}
	now := time.Now()
	rows := make([]model.WipWaferHistory, 0, len(wafers))
	for _, wafer := range wafers {
		rows = append(rows, model.WipWaferHistory{
			WaferID: wafer.ID, LotID: lotID, MoveID: moveID, EventType: event,
			FromNodeKey: node, ToNodeKey: node, ToSlot: wafer.Slot, CarrierID: wafer.CarrierID, CreatedAt: &now,
		})
	}
	return tx.Create(&rows).Error
}

func scrapWaferCount(tx *gorm.DB, lotID uint64, n int, reason string, moveID uint64) error {
	if n <= 0 {
		return nil
	}
	wafers, err := activeWafers(tx, lotID)
	if err != nil || len(wafers) == 0 {
		return err
	}
	if len(wafers) < n {
		return ErrWipQty
	}
	ids := make([]uint64, 0, n)
	for _, wafer := range wafers[len(wafers)-n:] {
		ids = append(ids, wafer.ID)
	}
	if err := scrapWaferIDs(tx, lotID, ids, reason, moveID); err != nil {
		return err
	}
	return compactSlots(tx, lotID)
}

func completeLotWafers(tx *gorm.DB, lotID uint64) error {
	wafers, err := activeWafers(tx, lotID)
	if err != nil || len(wafers) == 0 {
		return err
	}
	now := time.Now()
	ids := make([]uint64, 0, len(wafers))
	rows := make([]model.WipWaferHistory, 0, len(wafers))
	for _, wafer := range wafers {
		ids = append(ids, wafer.ID)
		rows = append(rows, model.WipWaferHistory{
			WaferID: wafer.ID, LotID: lotID, EventType: model.EventComplete, ToSlot: wafer.Slot, CarrierID: wafer.CarrierID, CreatedAt: &now,
		})
	}
	if err := tx.Model(&model.WipWafer{}).Where("id IN ?", ids).Update("status", model.WaferCompleted).Error; err != nil {
		return err
	}
	return tx.Create(&rows).Error
}

func activeWafers(tx *gorm.DB, lotID uint64) ([]model.WipWafer, error) {
	var rows []model.WipWafer
	err := tx.Where("lot_id = ? AND status = ?", lotID, model.WaferActive).Order("slot ASC, id ASC").Find(&rows).Error
	return rows, err
}

func countActive(tx *gorm.DB, lotID uint64) (int64, error) {
	var n int64
	err := tx.Model(&model.WipWafer{}).Where("lot_id = ? AND status = ?", lotID, model.WaferActive).Count(&n).Error
	return n, err
}

func pickWafers(tx *gorm.DB, lotID uint64, qty int, ids []uint64) ([]model.WipWafer, error) {
	rows, err := activeWafers(tx, lotID)
	if err != nil {
		return nil, err
	}
	if len(ids) > 0 {
		want := map[uint64]struct{}{}
		for _, id := range ids {
			want[id] = struct{}{}
		}
		if len(want) != qty {
			return nil, ErrWipQty
		}
		picked := make([]model.WipWafer, 0, qty)
		for _, row := range rows {
			if _, ok := want[row.ID]; ok {
				picked = append(picked, row)
			}
		}
		if len(picked) != qty {
			return nil, ErrWipQty
		}
		return picked, nil
	}
	if len(rows) < qty {
		return nil, ErrWipQty
	}
	return rows[len(rows)-qty:], nil
}

func scrapWaferIDs(tx *gorm.DB, lotID uint64, ids []uint64, reason string, moveID uint64) error {
	var rows []model.WipWafer
	if err := tx.Where("lot_id = ? AND status = ? AND id IN ?", lotID, model.WaferActive, ids).Find(&rows).Error; err != nil {
		return err
	}
	if len(rows) != len(ids) {
		return ErrWipQty
	}
	now := time.Now()
	for _, wafer := range rows {
		if err := tx.Model(&model.WipWafer{}).Where("id = ?", wafer.ID).Updates(map[string]any{
			"status": model.WaferScrapped, "slot": 0, "carrier_id": 0,
		}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.WipWaferHistory{
			WaferID: wafer.ID, LotID: lotID, MoveID: moveID, EventType: model.EventScrap,
			FromSlot: wafer.Slot, ReasonCode: reason, CarrierID: wafer.CarrierID, CreatedAt: &now,
		}).Error; err != nil {
			return err
		}
	}
	return compactSlots(tx, lotID)
}

func placeWafers(tx *gorm.DB, wafers []model.WipWafer, lotID, carrierID uint64, event string) error {
	now := time.Now()
	for i, wafer := range wafers {
		slot := i + 1
		if err := tx.Model(&model.WipWafer{}).Where("id = ?", wafer.ID).Updates(map[string]any{
			"carrier_id": carrierID, "slot": slot,
		}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.WipWaferHistory{
			WaferID: wafer.ID, LotID: lotID, EventType: event, FromSlot: wafer.Slot, ToSlot: slot, CarrierID: carrierID, CreatedAt: &now,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func compactSlots(tx *gorm.DB, lotID uint64) error {
	wafers, err := activeWafers(tx, lotID)
	if err != nil || len(wafers) == 0 {
		return err
	}
	carrierID := uint64(0)
	var bind model.WipCarrierBind
	err = tx.Where("lot_id = ? AND status = ?", lotID, model.BindBound).First(&bind).Error
	if err == nil {
		var carrier model.WipCarrier
		if err := tx.First(&carrier, bind.CarrierID).Error; err != nil {
			return err
		}
		if len(wafers) > carrier.Capacity {
			return ErrWipQty
		}
		carrierID = carrier.ID
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	for i, wafer := range wafers {
		slot := i + 1
		if wafer.Slot == slot && wafer.CarrierID == carrierID {
			continue
		}
		if err := tx.Model(&model.WipWafer{}).Where("id = ?", wafer.ID).Updates(map[string]any{
			"slot": slot, "carrier_id": carrierID,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}
