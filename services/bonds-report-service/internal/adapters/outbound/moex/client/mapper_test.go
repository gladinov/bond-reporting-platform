//go:build unit

package moex

import (
	"bonds-report-service/internal/adapters/outbound/moex/dto"
	"testing"

	domainModel "bonds-report-service/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestMapNullStringFromDTOToDomain(t *testing.T) {
	tests := []struct {
		name string
		dto  dto.NullString
		want domainModel.NullString
	}{
		{
			name: "value is set",
			dto: dto.NullString{
				Value:  "ABC",
				IsSet:  true,
				IsNull: false,
			},
			want: domainModel.NewNullString("ABC", true, false),
		},
		{
			name: "explicit null",
			dto: dto.NullString{
				Value:  "",
				IsSet:  true,
				IsNull: true,
			},
			want: domainModel.NewNullString("", true, true),
		},
		{
			name: "not set",
			dto: dto.NullString{
				Value:  "",
				IsSet:  false,
				IsNull: false,
			},
			want: domainModel.NewNullString("", false, false),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapNullStringFromDTOToDomain(tt.dto)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMapNullFloat64FromDTOToDomain(t *testing.T) {
	tests := []struct {
		name string
		dto  dto.NullFloat64
		want domainModel.NullFloat64
	}{
		{
			name: "value is set",
			dto: dto.NullFloat64{
				Value:  12.34,
				IsSet:  true,
				IsNull: false,
			},
			want: domainModel.NewNullFloat64(12.34, true, false),
		},
		{
			name: "explicit null",
			dto: dto.NullFloat64{
				Value:  0,
				IsSet:  true,
				IsNull: true,
			},
			want: domainModel.NewNullFloat64(0, true, true),
		},
		{
			name: "not set",
			dto: dto.NullFloat64{
				Value:  0,
				IsSet:  false,
				IsNull: false,
			},
			want: domainModel.NewNullFloat64(0, false, false),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapNullFloat64FromDTOToDomain(tt.dto)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMapValueFromDTOToDomain(t *testing.T) {
	dto := dto.Values{
		ShortName: dto.NullString{
			Value:  "OFZ 26238",
			IsSet:  true,
			IsNull: false,
		},
		TradeDate: dto.NullString{
			Value:  "2024-01-01",
			IsSet:  true,
			IsNull: false,
		},
		MaturityDate: dto.NullString{
			Value:  "2034-01-01",
			IsSet:  true,
			IsNull: false,
		},
		OfferDate: dto.NullString{
			Value:  "",
			IsSet:  true,
			IsNull: true,
		},
		BuybackDate: dto.NullString{
			IsSet:  false,
			IsNull: false,
		},
		YieldToMaturity: dto.NullFloat64{
			Value:  9.87,
			IsSet:  true,
			IsNull: false,
		},
		YieldToOffer: dto.NullFloat64{
			IsSet:  false,
			IsNull: false,
		},
		FaceValue: dto.NullFloat64{
			Value:  1000,
			IsSet:  true,
			IsNull: false,
		},
		FaceUnit: dto.NullString{
			Value:  "RUB",
			IsSet:  true,
			IsNull: false,
		},
		Duration: dto.NullFloat64{
			Value:  4.56,
			IsSet:  true,
			IsNull: false,
		},
	}

	got := MapValueFromDTOToDomain(dto)

	want := domainModel.ValuesMoex{
		ShortName:       domainModel.NewNullString("OFZ 26238", true, false),
		TradeDate:       domainModel.NewNullString("2024-01-01", true, false),
		MaturityDate:    domainModel.NewNullString("2034-01-01", true, false),
		OfferDate:       domainModel.NewNullString("", true, true),
		BuybackDate:     domainModel.NewNullString("", false, false),
		YieldToMaturity: domainModel.NewNullFloat64(9.87, true, false),
		YieldToOffer:    domainModel.NewNullFloat64(0, false, false),
		FaceValue:       domainModel.NewNullFloat64(1000, true, false),
		FaceUnit:        domainModel.NewNullString("RUB", true, false),
		Duration:        domainModel.NewNullFloat64(4.56, true, false),
	}

	assert.Equal(t, want, got)
}
