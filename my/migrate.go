//メインパッケージを定義
package my

//各種パッケージをimport
import (
        "fmt"

        "github.com/jinzhu/gorm"
        _ "github.com/jinzhu/gorm/dialects/sqlite"
)

// Migrate program.
func Migrate() {

        //データベースの接続
        db, er := gorm.Open("sqlite3", "data.sqlite3")

        //データベースのマイグレーション
        if er != nil {
                fmt.Println(er)
                return
        }
        defer db.Close()

        //データベースのマイグレーション
        db.AutoMigrate(&User{}, &Group{}, &Post{}, &Comment{})

}


