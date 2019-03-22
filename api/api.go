package api

import (
	context "context"
	"errors"
	"log"

	"github.com/dsociative/incrementer/db"
)

var (
	errMaximumBelowZero    = errors.New("maximum can't be less or equal zero")
	errStepLessOrEqualZero = errors.New("step can't be less or equal zero")
	empty                  = &Empty{}
)

type API struct {
	db db.DB
}

func NewAPI(db db.DB) *API {
	return &API{db: db}
}

func logIfError(method string, err error) error {
	if err != nil {
		log.Printf("method:%s error:%s", method, err)
	}
	return err
}

func (a *API) GetNumber(ctx context.Context, _ *Empty) (*Number, error) {
	number, err := a.db.Number()
	return &Number{Number: int64(number)}, logIfError("GetNumber", err)
}

func (a *API) IncrementNumber(ctx context.Context, _ *Empty) (*Number, error) {
	number, err := a.db.Incr()
	return &Number{Number: int64(number)}, logIfError("IncrementNumber", err)
}

func (a *API) SetSettings(ctx context.Context, setting *Setting) (*Empty, error) {
	if setting.Maximum <= 0 {
		return empty, errMaximumBelowZero
	} else if setting.Step <= 0 {
		return empty, errStepLessOrEqualZero
	}
	return empty, logIfError(
		"SetSettings",
		a.db.SetSettings(int(setting.Maximum), int(setting.Step)),
	)
}
