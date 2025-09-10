package middleware

import (
	"context"
	"strings"

	"github.com/cheel98/flashcard-backend/pkg/jwt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthMiddleware JWT认证中间件
type AuthMiddleware struct {
	jwtManager *jwt.JWTManager
	logger     *zap.Logger
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(jwtManager *jwt.JWTManager, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
		logger:     logger,
	}
}

// UnaryInterceptor 一元RPC拦截器
func (a *AuthMiddleware) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 检查是否需要认证
		if a.isPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		// 验证token
		claims, err := a.authorize(ctx)
		if err != nil {
			a.logger.Error("认证失败", zap.String("method", info.FullMethod), zap.Error(err))
			return nil, err
		}

		// 将用户信息添加到上下文
		ctx = a.addUserToContext(ctx, claims)

		return handler(ctx, req)
	}
}

// StreamInterceptor 流式RPC拦截器
func (a *AuthMiddleware) StreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// 检查是否需要认证
		if a.isPublicMethod(info.FullMethod) {
			return handler(srv, ss)
		}

		// 验证token
		claims, err := a.authorize(ss.Context())
		if err != nil {
			a.logger.Error("认证失败", zap.String("method", info.FullMethod), zap.Error(err))
			return err
		}

		// 创建包装的流，添加用户信息到上下文
		wrappedStream := &wrappedServerStream{
			ServerStream: ss,
			ctx:          a.addUserToContext(ss.Context(), claims),
		}

		return handler(srv, wrappedStream)
	}
}

// authorize 验证token
func (a *AuthMiddleware) authorize(ctx context.Context) (*jwt.Claims, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "缺少metadata")
	}

	values := md["authorization"]
	if len(values) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "缺少authorization header")
	}

	accessToken := values[0]
	if !strings.HasPrefix(accessToken, "Bearer ") {
		return nil, status.Errorf(codes.Unauthenticated, "无效的authorization header格式")
	}

	accessToken = strings.TrimPrefix(accessToken, "Bearer ")
	claims, err := a.jwtManager.VerifyToken(accessToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "无效的access token: %v", err)
	}

	// 检查token类型
	if claims.TokenType != jwt.AccessToken {
		return nil, status.Errorf(codes.Unauthenticated, "无效的token类型")
	}

	return claims, nil
}

// isPublicMethod 检查是否为公开方法（不需要认证）
func (a *AuthMiddleware) isPublicMethod(method string) bool {
	publicMethods := []string{
		"/user.UserService/Register",
		"/user.UserService/SendEmailCaptcha",
		"/user.UserService/VerifyCaptcha",
		"/user.UserService/Login",
		"/user.UserService/RefreshToken",
	}

	for _, publicMethod := range publicMethods {
		if method == publicMethod {
			return true
		}
	}
	return false
}

// addUserToContext 将用户信息添加到上下文
func (a *AuthMiddleware) addUserToContext(ctx context.Context, claims *jwt.Claims) context.Context {
	return context.WithValue(ctx, "user_id", claims.UserID)
}

// GetUserIDFromContext 从上下文获取用户ID
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value("user_id").(string)
	return userID, ok
}

// wrappedServerStream 包装的服务器流
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
