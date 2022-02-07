package bundlefx

import (
	"github.com/joeydtaylor/go-microservice/middleware/auth"
	"github.com/joeydtaylor/go-microservice/middleware/logger"
	"github.com/joeydtaylor/go-microservice/middleware/metrics"
	"go.uber.org/fx"
)

// Module provided to fx
var Module = fx.Options(
	auth.Module,
	logger.Module,
	metrics.Module,
)
