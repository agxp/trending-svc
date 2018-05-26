package main

import (
	"context"
	"database/sql"
	"github.com/agxp/cloudflix/video-hosting-svc/proto"
	"github.com/minio/minio-go"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"time"
)

type Repository interface {
	GetTrending(p opentracing.SpanContext) ([]*video_host.GetVideoInfoResponse, error)
	Prune(p opentracing.SpanContext) (uint64, error)
}

type TrendingRepository struct {
	s3     *minio.Client
	pg     *sql.DB
	tracer *opentracing.Tracer
	vh     video_host.HostClient
}

func (repo *TrendingRepository) GetTrending(parent opentracing.SpanContext) ([]*video_host.GetVideoInfoResponse, error) {
	sp, _ := opentracing.StartSpanFromContext(context.Background(), "GetTrending_Repo", opentracing.ChildOf(parent))
	defer sp.Finish()

	dbSP, _ := opentracing.StartSpanFromContext(context.Background(), "PG_GetTrending", opentracing.ChildOf(sp.Context()))
	defer dbSP.Finish()

	getTrendingQuery := `select id from videos where uploaded=true order by date_uploaded limit 20`

	rows, err := repo.pg.Query(getTrendingQuery)
	if err != nil {
		logger.Error(err.Error())
		dbSP.Finish()
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			logger.Error(err.Error())
			dbSP.Finish()
			return nil, err
		}
		logger.Info("trending id", zap.String("id", id))
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		logger.Error(err.Error())
		dbSP.Finish()
		return nil, err
	}

	var data []*video_host.GetVideoInfoResponse
	
	for _, v := range ids {
		res, err := repo.vh.GetVideoInfo(context.Background(), &video_host.GetVideoInfoRequest{
			Id: v,
		})
		
		if err != nil {
			logger.Error(err.Error())
			dbSP.Finish()
			return nil, err
		}
		data = append(data, res)
	}

	dbSP.Finish()

	return data, nil
}

func (repo *TrendingRepository) Prune(parent opentracing.SpanContext) (uint64, error) {
	sp, _ := opentracing.StartSpanFromContext(context.Background(), "Prune_Repo", opentracing.ChildOf(parent))
	defer sp.Finish()

	dbSP, _ := opentracing.StartSpanFromContext(context.Background(), "PG_Prune", opentracing.ChildOf(sp.Context()))
	defer dbSP.Finish()

	now := time.Now()
	sp.LogKV("now", now)

	pruneQuery := `delete from videos where uploaded=false and timeout_date < $1`

	r, err := repo.pg.Exec(pruneQuery, now)
	if err != nil {
		logger.Error(err.Error())
		dbSP.Finish()
		return 0, err
	}
	nRows, err := r.RowsAffected()
	if err != nil {
		logger.Error(err.Error())
		dbSP.Finish()
		return 0, err
	}
	sp.LogKV("rowsPruned", nRows)
	return uint64(nRows), nil
}
