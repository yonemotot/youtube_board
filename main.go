// package main

// import (
//         "./my"

//         _ "github.com/jinzhu/gorm/dialects/sqlite"
// )

// // main program.
//マイグレートを行う。
// func main() {
//         my.Migrate()
// }

//メインパッケージを定義
package main

//各種パッケージをimport
import (
        "log"
        "net/http"
        "strconv"
        "strings"
        "text/template"

        "./my"
        "github.com/gorilla/sessions"

        "github.com/jinzhu/gorm"
        _ "github.com/jinzhu/gorm/dialects/sqlite"
)

// db variable.
//SQLite3 データベースのドライバを指定するための変数
// SQLite3 データベースのファイル名や識別子を指定するための変数
var dbDriver = "sqlite3"
var dbName = "data.sqlite3"

// session variable.
//Goの github.com/gorilla/sessions パッケージを使用して、ウェブアプリケーション内でセッションを管理するための基本的な設定を行う。
//変数はセッションのクッキー名を指定
//cs 変数は sessions.NewCookieStore 関数を使用して新しいセッション用のクッキーストアを生成している。このクッキーストアはセッションデータをクッキーに保存し、クライアントとサーバー間でセッションを管理するために使用する。
var sesName = "ytboard-session"
var cs = sessions.NewCookieStore([]byte("secret-key-1234"))

// login check.
//ウェブアプリケーションにおいてユーザーのログイン状態をチェックし、ログインしていない場合にはログインページにリダイレクトするためのもの。

func checkLogin(w http.ResponseWriter, rq *http.Request) *my.User {
        //セッションの取得:cs.Get(rq, sesName) を使って、クッキーからセッションを取得する。
        ses, _ := cs.Get(rq, sesName)
        
        //ログインのチェック:セッションに "login" というキーが存在し、その値が true であればログイン済みと判定する。ログインしていない場合は、http.Redirect を使用してログインページにリダイレクトする。
        if ses.Values["login"] == nil || !ses.Values["login"].(bool) {
                http.Redirect(w, rq, "/login", 302)
        }
        //アカウント情報の取得:セッションから "account" というキーで保存されているアカウント情報を取得する。
        ac := ""
        if ses.Values["account"] != nil {
                ac = ses.Values["account"].(string)
        }

        //データベースからユーザー情報の取得:アカウント情報を元にデータベースから対応するユーザー情報を取得する。この際、GORM（ORMライブラリ）が使用されている。
        var user my.User
        db, _ := gorm.Open(dbDriver, dbName)
        defer db.Close()

        //ユーザー情報の返却:取得したユーザー情報をポインタとして返す。
        db.Where("account = ?", ac).First(&user)

        return &user
}

// Template for no-template.
//この関数は、何らかの理由でテンプレートを正しく生成できない場合に、代替として "NO PAGE." という文字列を表示するための簡単なテンプレートを生成する。
func notemp() *template.Template {
        tmp, _ := template.New("index").Parse("NO PAGE.")
        return tmp
}

// get target Temlate.
//この関数は、与えられたファイル名に対応するメインのテンプレートと、それに含まれる他の部分のテンプレートを結合して返す。
func page(fname string) *template.Template {
        tmps, _ := template.ParseFiles("templates/"+fname+".html",
                "templates/head.html", "templates/foot.html")
        return tmps
}

// top page handler.
func index(w http.ResponseWriter, rq *http.Request) {
        //ログインチェック:checkLogin 関数を使用して、ユーザーがログインしているかどうかをチェックし、ログインしている場合はユーザー情報を取得する。
        user := checkLogin(w, rq)

        // データベースに接続
        db, _ := gorm.Open(dbDriver, dbName)
        defer db.Close()

         // 直近の10件のグループへの投稿を取得
        var pl []my.Post
        db.Where("group_id > 0").Order("created_at desc").Limit(10).Find(&pl)
        
        // 直近の10件のグループを取得
        var gl []my.Group
        db.Order("created_at desc").Limit(10).Find(&gl)


         // 表示用のデータを構造体にまとめる
        item := struct {
                Title   string
                Message string
                Name    string
                Account string
                Plist   []my.Post
                Glist   []my.Group
        }{
                Title:   "Index",
                Message: "This is Top page.",
                Name:    user.Name,
                Account: user.Account,
                Plist:   pl,
                Glist:   gl,
        }

        // テンプレートにデータを渡してHTMLを生成し、レスポンスに書き込む
        er := page("index").Execute(w, item)
        if er != nil {
                log.Fatal(er)
        }
}

// top page handler.
func post(w http.ResponseWriter, rq *http.Request) {

        // ログインチェックを行い、ログインしているユーザー情報を取得
        user := checkLogin(w, rq)

        // URLのクエリパラメータから投稿IDを取得
        pid := rq.FormValue("pid")

        // データベースに接続
        db, _ := gorm.Open(dbDriver, dbName)
        defer db.Close()

        // POSTリクエストがあれば、コメントをデータベースに追加
        if rq.Method == "POST" {
                msg := rq.PostFormValue("message")
                pId, _ := strconv.Atoi(pid)
                cmt := my.Comment{
                        UserId:  int(user.Model.ID),
                        PostId:  pId,
                        Message: msg,
                }
                db.Create(&cmt)
        }

        // 投稿とそのコメントを取得
        var pst my.Post
        var cmts []my.CommentJoin

        db.Where("id = ?", pid).First(&pst)
        db.Table("comments").Select("comments.*, users.id, users.name").Joins("join users on users.id =comments.user_id").Where("comments.post_id = ?", pid).Order("created_at desc").Find(&cmts)

        // 表示用のデータを構造体にまとめる
        item := struct {
                Title   string
                Message string
                Name    string
                Account string
                Post    my.Post
                Clist   []my.CommentJoin
        }{
                Title:   "Post",
                Message: "Post id=" + pid,
                Name:    user.Name,
                Account: user.Account,
                Post:    pst,
                Clist:   cmts,
        }

        // テンプレートにデータを渡してHTMLを生成し、レスポンスに書き込む
        er := page("post").Execute(w, item)
        if er != nil {
                log.Fatal(er)
        }
}

// home handler
func home(w http.ResponseWriter, rq *http.Request) {

        // ログインチェックを行い、ログインしているユーザー情報を取得
        user := checkLogin(w, rq)

        // データベースに接続
        db, _ := gorm.Open(dbDriver, dbName)
        defer db.Close()

        // POSTリクエストがあれば、フォームの内容に応じて投稿またはグループをデータベースに追加
        if rq.Method == "POST" {
                switch rq.PostFormValue("form") {
                case "post":
                        ad := rq.PostFormValue("address")
                        ad = strings.TrimSpace(ad)
                        if strings.HasPrefix(ad, "https://youtu.be/") {
                                ad = strings.TrimPrefix(ad, "https://youtu.be/")
                        }

                        // ユーザーの投稿をデータベースに追加
                        pt := my.Post{
                                UserId:  int(user.Model.ID),
                                Address: ad,
                                Message: rq.PostFormValue("message"),
                        }
                        db.Create(&pt)
                case "group":

                        // ユーザーのグループをデータベースに追加
                        gp := my.Group{
                                UserId:  int(user.Model.ID),
                                Name:    rq.PostFormValue("name"),
                                Message: rq.PostFormValue("message"),
                        }
                        db.Create(&gp)
                }
        }

        // ユーザーの直近の10件の投稿とグループを取得
        var pts []my.Post
        var gps []my.Group

        db.Where("user_id=?", user.ID).Order("created_at desc").Limit(10).Find(&pts)
        db.Where("user_id=?", user.ID).Order("created_at desc").Limit(10).Find(&gps)

        // 表示用のデータを構造体にまとめる
        itm := struct {
                Title   string
                Message string
                Name    string
                Account string
                Plist   []my.Post
                Glist   []my.Group
        }{
                Title:   "Home",
                Message: "User account=\"" + user.Account + "\".",
                Name:    user.Name,
                Account: user.Account,
                Plist:   pts,
                Glist:   gps,
        }

         // テンプレートにデータを渡してHTMLを生成し、レスポンスに書き込む
        er := page("home").Execute(w, itm)
        if er != nil {
                log.Fatal(er)
        }
}

// group handler.
func group(w http.ResponseWriter, rq *http.Request) {

        // ログインチェックを行い、ログインしているユーザー情報を取得
        user := checkLogin(w, rq)

        // URLのクエリパラメータからグループIDを取得
        gid := rq.FormValue("gid")

        // データベースに接続
        db, _ := gorm.Open(dbDriver, dbName)
        defer db.Close()

        // POSTリクエストがあれば、フォームの内容に応じてグループに投稿をデータベースに追加
        if rq.Method == "POST" {
                ad := rq.PostFormValue("address")
                ad = strings.TrimSpace(ad)
                if strings.HasPrefix(ad, "https://youtu.be/") {
                        ad = strings.TrimPrefix(ad, "https://youtu.be/")
                }
                gId, _ := strconv.Atoi(gid)

                // ユーザーが特定のグループに投稿したデータをデータベースに追加
                pt := my.Post{
                        UserId:  int(user.Model.ID),
                        Address: ad,
                        Message: rq.PostFormValue("message"),
                        GroupId: gId,
                }
                db.Create(&pt)
        }

        // グループ情報とそのグループに関連する投稿を取得
        var grp my.Group
        var pts []my.Post

        db.Where("id=?", gid).First(&grp)
        db.Order("created_at desc").Model(&grp).Related(&pts)

        // 表示用のデータを構造体にまとめる
        itm := struct {
                Title   string
                Message string
                Name    string
                Account string
                Group   my.Group
                Plist   []my.Post
        }{
                Title:   "Group",
                Message: "Group id=" + gid,
                Name:    user.Name,
                Account: user.Account,
                Group:   grp,
                Plist:   pts,
        }

        // テンプレートにデータを渡してHTMLを生成し、レスポンスに書き込む
        er := page("group").Execute(w, itm)
        if er != nil {
                log.Fatal(er)
        }
}

// login handler.
func login(w http.ResponseWriter, rq *http.Request) {

        // ログイン情報を格納する構造体を作成
        item := struct {
                Title   string
                Message string
                Account string
        }{
                Title:   "Login",
                Message: "type your account & password:",
                Account: "",
        }

        // GETリクエストの場合はログインページを表示
        if rq.Method == "GET" {
                er := page("login").Execute(w, item)
                if er != nil {
                        log.Fatal(er)
                }
                return
        }

        // POSTリクエストの場合はログイン処理を行う
        if rq.Method == "POST" {

                // データベースに接続
                db, _ := gorm.Open(dbDriver, dbName)
                defer db.Close()

                // フォームから入力されたアカウントとパスワードを取得
                usr := rq.PostFormValue("account")
                pass := rq.PostFormValue("pass")
                item.Account = usr

                // check account and password
                var re int
                var user my.User

                 // データベースからアカウントとパスワードが一致するユーザーを検索
                db.Where("account = ? and password = ?", usr, pass).Find(&user).Count(&re)

                if re <= 0 {

                        // 一致するユーザーがいない場合はエラーメッセージを表示
                        item.Message = "Wrong account or password."
                        page("login").Execute(w, item)
                        return
                }

                // logined.
                // ログイン成功時の処理
                ses, _ := cs.Get(rq, sesName)
                ses.Values["login"] = true
                ses.Values["account"] = usr
                ses.Values["name"] = user.Name
                ses.Save(rq, w)
                // ログイン成功後、ホームページにリダイレクト
                http.Redirect(w, rq, "/", 302)
        }

        // ログインページを表示
        er := page("login").Execute(w, item)
        if er != nil {
                log.Fatal(er)
        }
}

// logout handler.
func logout(w http.ResponseWriter, rq *http.Request) {

        // セッションからログイン情報をクリア
        ses, _ := cs.Get(rq, sesName)
        ses.Values["login"] = nil
        ses.Values["account"] = nil
        ses.Save(rq, w)

        // ログインページにリダイレクト
        http.Redirect(w, rq, "/login", 302)
}

// main program.
func main() {
        // index handling.
        http.HandleFunc("/", func(w http.ResponseWriter, rq *http.Request) {
                index(w, rq)
        })
        // home handling.
        http.HandleFunc("/home", func(w http.ResponseWriter, rq *http.Request) {
                home(w, rq)
        })
        // post handling.
        http.HandleFunc("/post", func(w http.ResponseWriter, rq *http.Request) {
                post(w, rq)
        })
        // post handling.
        http.HandleFunc("/group", func(w http.ResponseWriter, rq *http.Request) {
                group(w, rq)
        })

        // login handling.
        http.HandleFunc("/login", func(w http.ResponseWriter, rq *http.Request) {
                login(w, rq)
        })
        // logout handling.
        http.HandleFunc("/logout", func(w http.ResponseWriter, rq *http.Request) {
                logout(w, rq)
        })

        http.ListenAndServe("", nil)
}
