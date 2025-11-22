package user

import (
	"blog/internal/common/constants"
	"blog/internal/config"
	errors "errors"
	"time"

	"gorm.io/gorm"

	"crypto/rand"
	"encoding/hex"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	UserRepository *UserRepository
}

func NewUserService(userRepository *UserRepository) *UserService {
	return &UserService{
		UserRepository: userRepository,
	}
}

func (s *UserService) Register(userInfo *CreateUserDTO) (*User, error) {
	findUser, err := s.UserRepository.FindByUserName(userInfo.UserName)
	if findUser != nil {
		return nil, errors.New("this username has already been registered")
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	salt := generateSalt()
	hashedPassword := hashPassword(userInfo.Password, salt)

	user := &User{
		UserName:  userInfo.UserName,
		Password:  hashedPassword,
		Salt:      salt,
		ID:        uuid.New().String(),
		Role:      userInfo.Role,
		Email:     userInfo.Email,
		Phone:     userInfo.FullName,
		Avatar:    userInfo.Avatar,
		Gender:    userInfo.Gender,
		FullName:  userInfo.FullName,
		Status:    constants.Active,
		Birthday:  userInfo.Birthday,
		Address:   userInfo.Address,
		CreatedAt: time.Now(),
	}

	return user, s.UserRepository.Create(user)
}

func (s *UserService) Authenticate(userName string, password string) (*User, string, string, error) {
	user, err := s.UserRepository.FindByUserName(userName)
	if user == nil {
		return nil, "", "", errors.New("the current username is not registered")
	}
	if err != nil {
		return nil, "", "", err
	}
	if !verifyPassword(password, user.Password, user.Salt) {
		return nil, "", "", errors.New("password error")
	}

	accessToken, err := generateAccessToken(user.UserName, user.Password)
	if err != nil {
		return nil, "", "", err
	}

	refreshToken, err := generateRefreshToken(user.UserName, user.Password)
	if err != nil {
		return nil, "", "", err
	}
	return user, accessToken, refreshToken, nil
}

func (s *UserService) RefreshToken(refreshToken string) (string, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return "", err
	}
	token, err := jwt.ParseWithClaims(refreshToken, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.SecretKey), nil
	})
	if err != nil {
		return "", err
	}
	if !token.Valid {
		return "", errors.New("refreshToken is invalid")
	}
	claims := token.Claims.(*jwt.StandardClaims)
	userName := claims.Subject
	user, err := s.UserRepository.FindByUserName(userName)
	if err != nil {
		return "", err
	}
	accessToken, err := generateAccessToken(user.UserName, user.Password)
	if err != nil {
		return "", err
	}
	return accessToken, nil
}

func generateSalt() string {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(salt)
}

func hashPassword(password, salt string) string {
	combined := []byte(password + salt)
	hashedPassword, _ := bcrypt.GenerateFromPassword(combined, bcrypt.DefaultCost)
	return string(hashedPassword)
}

func verifyPassword(inputPassword, hashedPassword, salt string) bool {
	combined := []byte(inputPassword + salt)
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), combined)
	return err == nil
}

func generateToken(userName string, hashedPassword string, duration time.Duration) (string, error) {
	// 设置 token 的过期时间为 7 天
	expirationTime := time.Now().Add(duration)
	claims := &jwt.StandardClaims{
		ExpiresAt: expirationTime.Unix(),
		Subject:   userName + hashedPassword,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err.Error())
	}
	return token.SignedString([]byte(cfg.SecretKey))
}

func generateAccessToken(username string, pwd string) (string, error) {
	accessToken, err := generateToken(username, pwd, 15*time.Minute)
	if err != nil {
		return "", err
	}
	return accessToken, err
}

func generateRefreshToken(username string, pwd string) (string, error) {
	refreshToken, err := generateToken(username, pwd, 7*24*time.Hour)
	if err != nil {
		return "", err
	}
	return refreshToken, err
}
