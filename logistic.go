package main

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DaftarUser = []User{}

type User struct {
	gorm.Model
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
	Alamat   string `json:"alamat" form:"alamat"`
}

var DaftarVendor = []Vendor{}

type Vendor struct {
	gorm.Model
	Name        string `json:"name" form:"name"`
	Category    string `json:"category" form:"category"`
	Hp          string `json:"hp" form:"hp"`
	VehicleType string `json:"vehicle_type" form:"vehicle_type"`
}

func connectDB() *gorm.DB {
	dsn := "root:@tcp(localhost:3306)/logistic?charset=utf8mb4&parseTime=True&loc=Local"
	db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	return db
}

// Login
func GetLogin(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {

		var loginData User
		if err := c.Bind(&loginData); err != nil {
			log.Error(err.Error())
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": "cannot process data",
			})
		}

		if err := db.First(&loginData, "username = ? && password =?", loginData.Username, loginData.Password).Error; err != nil {
			log.Error(err.Error())
			c.JSON(http.StatusNotFound, map[string]interface{}{
				"message": "cannot find any data",
			})
		}
		tkn := GenerateToken(loginData.ID)
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "success on login",
			"data":    loginData,
			"token":   tkn,
		})
	}
}

// Register
func PostRegister(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var input User
		if err := c.Bind(&input); err != nil {
			log.Error(err)
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": "cannot read data",
			})
		}

		err := db.Create(&input).Error
		if err != nil {
			log.Error(err)
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"message": "cannot insert data",
			})
		}

		return c.JSON(http.StatusCreated, map[string]interface{}{
			"message": "success insert new user",
			"data":    input,
		})
	}
}

// Get All Data Vendor
func AllVendor(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var qryRes []Vendor
		if err := db.Find(&qryRes).Error; err != nil {
			log.Error(err.Error())
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"message": "error on database",
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "success get all data",
			"data":    qryRes,
		})
	}
}

// Data Vendor
func DateVendor(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		category := c.Param("category")
		var qryRes []Vendor
		if err := db.Where("category = ?", category).Find(&qryRes).Error; err != nil {
			log.Error(err.Error())
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"message": "error on database",
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "success get specific data",
			"data":    qryRes,
		})
	}
}

// Tambah Data Vendor (POST)
func CreateVendor(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var input Vendor
		if err := c.Bind(&input); err != nil {
			log.Error(err)
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": "cannot read data",
			})
		}

		if err := db.Create(&input).Error; err != nil {
			log.Error(err)
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"message": "cannot insert data",
			})
		}

		return c.JSON(http.StatusCreated, map[string]interface{}{
			"message": "success insert new Vendor",
			"data":    input,
		})
	}
}

// Token JWT
func GenerateToken(id uint) string {
	claim := make(jwt.MapClaims)
	claim["authorized"] = true
	claim["id"] = id
	claim["exp"] = time.Now().Add(time.Hour * 1).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)

	str, err := token.SignedString([]byte("kaguya-sama"))
	if err != nil {
		log.Error(err.Error())
		return ""
	}
	return str
}

// Extract Token
func ExtractToken(c echo.Context) uint {
	token := c.Get("vendor").(*jwt.Token)
	if token.Valid {
		claim := token.Claims.(jwt.MapClaims)
		return uint(claim["hp"].(float64))
	}
	return 0
}

func main() {
	e := echo.New()
	db := connectDB()
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Vendor{})

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}, time=${time_rfc3339_nano}\n ",
	}))
	e.Use(middleware.JWT([]byte("Alif123")))

	o := e.Group("/orm")
	o.GET("/login", GetLogin(db))
	o.POST("/users", PostRegister(db))
	o.GET("/vendor", AllVendor(db))
	o.GET("/dateVendor/:category", DateVendor(db), middleware.BasicAuth(func(username, password string, ctx echo.Context) (bool, error) {
		user := User{}
		if err := db.First(&user, "username = ? AND password = ?", username, password); err.Error != nil {
			log.Error(user)
			return false, err.Error
		}
		return true, nil
	}))
	o.POST("/vendors", CreateVendor(db), middleware.JWT([]byte("kaguya-sama")))
	e.Start(":8000")
}
