package handler

import (
	"fmt"
	"math"

	db "github.com/MamangRust/microservice-ecommerce-grpc-banner/database/schema"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/banner"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/common"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return page, pageSize
}

func createPaginationMeta(page, pageSize, totalRecords int) *pb_common.PaginationMeta {
	totalPages := int(math.Ceil(float64(totalRecords) / float64(pageSize)))
	return &pb_common.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(totalRecords),
	}
}

func formatTimestamp(v interface{}) string {
	if ts, ok := v.(pgtype.Timestamptz); ok && ts.Valid {
		return ts.Time.Format("2006-01-02 15:04:05.000")
	}
	return ""
}

func formatDate(v interface{}) (string, bool) {
	if d, ok := v.(pgtype.Date); ok && d.Valid {
		return d.Time.Format("2006-01-02"), true
	}
	return "", false
}

func formatTime(v interface{}) (string, bool) {
	if t, ok := v.(pgtype.Time); ok && t.Valid {
		hours := t.Microseconds / (1000000 * 60 * 60)
		minutes := (t.Microseconds / (1000000 * 60)) % 60
		seconds := (t.Microseconds / 1000000) % 60
		return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds), true
	}
	return "", false
}

func mapToProtoBannerResponse(m interface{}) *pb_banner.BannerResponse {
	var res *pb_banner.BannerResponse

	switch v := m.(type) {
	case *db.Banner:
		res = &pb_banner.BannerResponse{
			BannerId:  int32(v.BannerID),
			Name:      v.Name,
			IsActive:  *v.IsActive,
			CreatedAt: formatTimestamp(v.CreatedAt),
			UpdatedAt: formatTimestamp(v.UpdatedAt),
		}
		if val, ok := formatDate(v.StartDate); ok {
			res.StartDate = val
		}
		if val, ok := formatDate(v.EndDate); ok {
			res.EndDate = val
		}
		if val, ok := formatTime(v.StartTime); ok {
			res.StartTime = val
		}
		if val, ok := formatTime(v.EndTime); ok {
			res.EndTime = val
		}
	case *db.GetBannersRow:
		res = &pb_banner.BannerResponse{
			BannerId:  v.BannerID,
			Name:      v.Name,
			IsActive:  *v.IsActive,
			CreatedAt: formatTimestamp(v.CreatedAt),
			UpdatedAt: formatTimestamp(v.UpdatedAt),
		}
		if val, ok := formatDate(v.StartDate); ok {
			res.StartDate = val
		}
		if val, ok := formatDate(v.EndDate); ok {
			res.EndDate = val
		}
		if val, ok := formatTime(v.StartTime); ok {
			res.StartTime = val
		}
		if val, ok := formatTime(v.EndTime); ok {
			res.EndTime = val
		}
	case *db.GetBannerRow:
		res = &pb_banner.BannerResponse{
			BannerId:  v.BannerID,
			Name:      v.Name,
			IsActive:  *v.IsActive,
			CreatedAt: formatTimestamp(v.CreatedAt),
			UpdatedAt: formatTimestamp(v.UpdatedAt),
		}
		if val, ok := formatDate(v.StartDate); ok {
			res.StartDate = val
		}
		if val, ok := formatDate(v.EndDate); ok {
			res.EndDate = val
		}
		if val, ok := formatTime(v.StartTime); ok {
			res.StartTime = val
		}
		if val, ok := formatTime(v.EndTime); ok {
			res.EndTime = val
		}
	case *db.CreateBannerRow:
		res = &pb_banner.BannerResponse{
			BannerId:  v.BannerID,
			Name:      v.Name,
			IsActive:  *v.IsActive,
			CreatedAt: formatTimestamp(v.CreatedAt),
			UpdatedAt: formatTimestamp(v.UpdatedAt),
		}
		if val, ok := formatDate(v.StartDate); ok {
			res.StartDate = val
		}
		if val, ok := formatDate(v.EndDate); ok {
			res.EndDate = val
		}
		if val, ok := formatTime(v.StartTime); ok {
			res.StartTime = val
		}
		if val, ok := formatTime(v.EndTime); ok {
			res.EndTime = val
		}
	case *db.UpdateBannerRow:
		res = &pb_banner.BannerResponse{
			BannerId:  v.BannerID,
			Name:      v.Name,
			IsActive:  *v.IsActive,
			CreatedAt: formatTimestamp(v.CreatedAt),
			UpdatedAt: formatTimestamp(v.UpdatedAt),
		}
		if val, ok := formatDate(v.StartDate); ok {
			res.StartDate = val
		}
		if val, ok := formatDate(v.EndDate); ok {
			res.EndDate = val
		}
		if val, ok := formatTime(v.StartTime); ok {
			res.StartTime = val
		}
		if val, ok := formatTime(v.EndTime); ok {
			res.EndTime = val
		}
	default:
		return nil
	}

	return res
}

func mapToProtoBannerResponseDeleteAt(m interface{}) *pb_banner.BannerResponseDeleteAt {
	var res *pb_banner.BannerResponseDeleteAt

	switch v := m.(type) {
	case *db.Banner:
		res = &pb_banner.BannerResponseDeleteAt{
			BannerId:  int32(v.BannerID),
			Name:      v.Name,
			IsActive:  *v.IsActive,
			CreatedAt: formatTimestamp(v.CreatedAt),
			UpdatedAt: formatTimestamp(v.UpdatedAt),
		}
		if val, ok := formatDate(v.StartDate); ok {
			res.StartDate = val
		}
		if val, ok := formatDate(v.EndDate); ok {
			res.EndDate = val
		}
		if val, ok := formatTime(v.StartTime); ok {
			res.StartTime = val
		}
		if val, ok := formatTime(v.EndTime); ok {
			res.EndTime = val
		}
		if val := formatTimestamp(v.DeletedAt); val != "" {
			res.DeletedAt = &wrapperspb.StringValue{Value: val}
		}
	case *db.GetBannersActiveRow:
		res = &pb_banner.BannerResponseDeleteAt{
			BannerId:  v.BannerID,
			Name:      v.Name,
			IsActive:  *v.IsActive,
			CreatedAt: formatTimestamp(v.CreatedAt),
			UpdatedAt: formatTimestamp(v.UpdatedAt),
		}
		if val, ok := formatDate(v.StartDate); ok {
			res.StartDate = val
		}
		if val, ok := formatDate(v.EndDate); ok {
			res.EndDate = val
		}
		if val, ok := formatTime(v.StartTime); ok {
			res.StartTime = val
		}
		if val, ok := formatTime(v.EndTime); ok {
			res.EndTime = val
		}
		if val := formatTimestamp(v.DeletedAt); val != "" {
			res.DeletedAt = &wrapperspb.StringValue{Value: val}
		}
	case *db.GetBannersTrashedRow:
		res = &pb_banner.BannerResponseDeleteAt{
			BannerId:  v.BannerID,
			Name:      v.Name,
			IsActive:  *v.IsActive,
			CreatedAt: formatTimestamp(v.CreatedAt),
			UpdatedAt: formatTimestamp(v.UpdatedAt),
		}
		if val, ok := formatDate(v.StartDate); ok {
			res.StartDate = val
		}
		if val, ok := formatDate(v.EndDate); ok {
			res.EndDate = val
		}
		if val, ok := formatTime(v.StartTime); ok {
			res.StartTime = val
		}
		if val, ok := formatTime(v.EndTime); ok {
			res.EndTime = val
		}
		if val := formatTimestamp(v.DeletedAt); val != "" {
			res.DeletedAt = &wrapperspb.StringValue{Value: val}
		}
	default:
		return nil
	}

	return res
}
