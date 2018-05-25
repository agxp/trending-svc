package main

import (
	pb "github.com/agxp/cloudflix/trending-svc/proto"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type service struct {
	repo   Repository
	tracer *opentracing.Tracer
	logger *zap.Logger
}

func (srv *service) GetTrending(ctx context.Context, req *pb.Request, res *pb.GetTrendingResponse) error {
	sp, _ := opentracing.StartSpanFromContext(ctx, "GetTrending_Service")

	logger.Info("Request for GetTrending_Service received")
	defer sp.Finish()

	rsp, err := srv.repo.GetTrending(sp.Context())
	if err != nil {
		logger.Error("failed GetTrending", zap.Error(err))
		return err
	}

	res.Data = rsp
	return nil
}


func (srv *service) Prune(ctx context.Context, req *pb.PruneRequest, res *pb.PruneResponse) error {
	sp, _ := opentracing.StartSpanFromContext(ctx, "Prune_Service")

	logger.Info("Request for Prune_Service received")
	defer sp.Finish()

	nRows, err := srv.repo.Prune(sp.Context())
	if err != nil {
		logger.Error("failed Prune", zap.Error(err))
		return err
	}

	res.NumPruned = nRows
	return nil
}