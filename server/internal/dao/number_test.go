package dao

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"semi-mes/server/internal/model"
)

func TestAllocateIsUniqueUnderConcurrency(t *testing.T) {
	dsn := fmt.Sprintf("file:number-race-%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(8)
	if err = db.AutoMigrate(&model.SysNumberRule{}, &model.SysNumberSeq{}); err != nil {
		t.Fatal(err)
	}
	if err = db.Create(&model.SysNumberRule{
		RuleCode: "lot", RuleName: "lot", Prefix: "L", DatePart: "none", SeqLength: 4, ResetPeriod: "never", Separator: "-", Status: 1,
	}).Error; err != nil {
		t.Fatal(err)
	}
	preview, err := PreviewNumber(db, "lot", time.Now())
	if err != nil || preview != "L-0001" {
		t.Fatalf("preview %s %v", preview, err)
	}
	const n = 20
	out := make(chan string, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			text, err := Allocate(db, "lot", time.Now())
			if err != nil {
				t.Errorf("allocate: %v", err)
				return
			}
			out <- text
		}()
	}
	wg.Wait()
	close(out)
	seen := map[string]struct{}{}
	for text := range out {
		if _, ok := seen[text]; ok {
			t.Fatalf("duplicate %s", text)
		}
		seen[text] = struct{}{}
	}
	if len(seen) != n {
		t.Fatalf("got %d numbers", len(seen))
	}
	again, err := PreviewNumber(db, "lot", time.Now())
	if err != nil || again != "L-0021" {
		t.Fatalf("next preview %s %v", again, err)
	}
}
