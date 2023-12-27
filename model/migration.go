package model

func migration() {
	DB.Set("gorm:table_options", "charset=utf8").AutoMigrate(&User{})
}
