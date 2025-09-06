package auth

// type AuthServiceServer struct {
// 	authpb.UnimplementedAuthServiceServer
// 	auth *AuthService
// }

// func NewAuthServiceServer(auth *AuthService) *AuthServiceServer {
// 	return &AuthServiceServer{auth: auth}
// }

// func (s *AuthServiceServer) CreateCredentials(ctx context.Context, req *authpb.CreateCredentialsRequest) (*authpb.CreateCredentialsResponse, error) {
// 	err := s.auth.CreateCredentials(ctx, req.UserId, req.Username, req.Password)
// 	if err != nil {
// 		return &authpb.CreateCredentialsResponse{Success: false}, err
// 	}
// 	return &authpb.CreateCredentialsResponse{Success: true}, nil
// }
// func (s *AuthServiceServer) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
// 	token, err := s.auth.Login(ctx, req.Username, req.Password)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &authpb.LoginResponse{
// 		UserId:      token.UserId,
// 		AccessToken: token.AccessToken,
// 		ExpiresIn:   int64(token.ExpiresIn),
// 	}, nil
// }

// // func (s *AuthServiceServer) ChangePassword(ctx context.Context, req *authpb.ChangePasswordRequest) (*authpb.ChangePasswordResponse, error) {
// // 	err := s.auth.ChangePassword(ctx, req.UserId, req.OldPassword, req.NewPassword)
// // 	if err != nil {
// // 		return &authpb.ChangePasswordResponse{Success: false}, err
// // 	}
// // 	return &authpb.ChangePasswordResponse{Success: true}, nil
// // }
