// internal/pkg/logger/business_log_helper.go
package logger

import (
	"context"

	"go.uber.org/zap"
)

// LogBusiness 记录一条业务日志，并可选镜像到系统日志
//
// 参数说明：
//   - ctx      : 上下文，用于提取 trace_id、ip 等公共字段（可由中间件注入到 ctx）。
//   - biz      : 业务日志记录器（通常是 app.App.BizLogger）。
//   - sys      : 系统日志记录器（通常是 app.App.Logger）。
//   - mirror   : 是否将业务日志同步写入系统日志（由配置决定）。
//   - action   : 本次业务操作的动作名称（如 "user.update"、"order.create"）。
//   - operator : 操作者（用户名、用户ID、client_id 等）。
//   - target   : 被操作对象（如 {"user_id":101}、{"order_id":999}）。
//   - details  : 补充信息（如更新前后差异、金额、请求参数等）。
//
// 功能：
//  1. 构造统一的日志字段（types=business, action, operator, trace_id, ip, target, details）。
//  2. 向业务日志（biz）写一条 Info 级别日志，便于审计与追踪。
//  3. 如果 mirror==true，则再向系统日志（sys）写一条同样内容的 Info，方便排错与统一检索。
//
// 使用场景：
//   - 在 Service 层或 Handler 中调用，记录用户操作、管理员行为、审计日志。
//   - 可配置业务日志是否双写到系统日志。
//   - 方便后续 ELK/Graylog 等日志平台做检索与分析。
func LogBusiness(
	ctx context.Context,
	biz *Logger,
	sys *Logger,
	mirror bool,
	action, operator string,
	target map[string]any,
	details map[string]any,
) {
	// 基础字段
	fields := []zap.Field{
		zap.String("types", "business"),
		zap.String("action", action),
		zap.String("operator", operator),
	}

	if ctx != nil {
		if v := ctx.Value("trace_id"); v != nil {
			if s, ok := v.(string); ok && s != "" {
				fields = append(fields, zap.String("trace_id", s))
			}
		}
		if v := ctx.Value("ip"); v != nil {
			if s, ok := v.(string); ok && s != "" {
				fields = append(fields, zap.String("ip", s))
			}
		}
	}

	// 操作对象（如 user_id / order_id）
	if target != nil {
		fields = append(fields, zap.Any("target", target))
	}
	// 详细信息（更新内容、金额等）
	if details != nil {
		fields = append(fields, zap.Any("details", details))
	}

	if biz != nil {
		biz.Info("business", fields...)
	}

	if mirror && sys != nil {
		sys.Info("business(mirror)", fields...)
	}
}
