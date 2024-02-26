package main

import (
	"NextGo/ent"
	"NextGo/ent/user"
	"context"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"log"
	"strconv"
)

func main() {
	//PostgreSQLに接続
	client, err := ent.Open("postgres", "host=db port=5432 user=postgres dbname=db password=password sslmode=disable")

	if err != nil {
		log.Fatalf("failed opening connection to postgres: %v", err)
	}
	// Run the auto migration tool.
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	//Ginフレームワークのデフォルトの設定を使用してルータを作成
	router := gin.Default()

	// CORS設定
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}
		c.Next()
	})

	// ルートハンドラの定義
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello, World!!!",
		})
	})

	// ユーザ新規登録機能
	router.POST("users/sign_up", func(c *gin.Context) {

		// サインアップで送られてくるリクエストを型定義
		type SignUpRequest struct {
			Username string `json:"username" binding:"required"`
			Email    string `json:"email" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		// 変数reqをSignUpRequestで定義
		var req SignUpRequest

		//reqに取得したデータを格納、変換でエラーが起きた場合はエラーを返して終了
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}

		// ユーザ登録を行う
		newUser, err := client.User.
			Create().
			SetUsername(req.Username).
			SetEmail(req.Email).
			SetPassword(req.Password).
			Save(context.Background())

		// エラーの場合はエラーを返して処理終了。
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error(), "messsage": "sign up missing"})
			return
		}
		// 保存したUserの情報をレスポンスとして返す。
		c.JSON(201, gin.H{"user": newUser})

	})

	// ユーザログイン機能
	router.POST("users/sign_in", func(c *gin.Context) {

		// ログインで送られてくるリクエストを型定義
		type SignInRequest struct {
			Email    string `json:"email" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		// 変数reqをSignInRequestで定義
		var req SignInRequest

		//reqに取得したデータを格納、変換でエラーが起きた場合はエラーを返して終了
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}

		// ユーザの検索を行う
		sign_in_user, err := client.User.Query().
			Where(user.EmailEQ(req.Email), user.PasswordEQ(req.Password)).
			First(context.Background())

		//エラーを返す
		if err != nil {
			c.JSON(401, gin.H{"error": "invalid credentials"})
			return
		}

		// ログイン成功
		c.JSON(200, gin.H{"user": sign_in_user})

	})

	// 本の新規登録
	router.POST("/books", func(c *gin.Context) {

		// 本の新規登録で送られてくるリクエストを型定義
		type NewBookRequest struct {
			Title  string `json:"title" binding:"required"`
			Author string `json:"author" binding:"required"`
			UserId int    `json:"user_id" binding:"required"`
		}

		// reqをNewBookRequestで定義
		var req NewBookRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}

		// 本の情報を保存
		newBook, err := client.Book.
			Create().
			SetTitle(req.Title).
			SetAuthor(req.Author).
			SetUserID(req.UserId).
			Save(context.Background())

		// エラーがある場合はエラーを返して終了
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error(), "message": "create book missing"})
			return
		}

		// 保存したBookの情報をレスポンスとして返す
		c.JSON(201, newBook)
	})

	// 本の一覧を取得
	router.GET("/books", func(c *gin.Context) {

		// Book一覧を取得する
		books, err := client.Book.Query().All(context.Background())

		// エラーならエラーを返して終了
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error(), "message": "Could not get the book list."})
			return
		}

		// booksをjson形式で返す
		c.JSON(200, books)
	})

	// 本の情報を取得
	router.GET("/books/:id", func(c *gin.Context) {

		// URLパラメータから本のIDを取得する。
		bookIDStr := c.Param("id")

		// 文字->数字変換
		bookID, err := strconv.Atoi(bookIDStr)

		// パラメータが不正な場合はエラーを出力して終了
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid Book ID"})
			return
		}

		// 指定されたIDの本をデータベースから検索
		book, err := client.Book.Get(context.Background(), bookID)

		// 本が見つからない場合はエラーを返して終了
		if err != nil {
			c.JSON(404, gin.H{"error": err.Error(), "message": "Book with specified id not found"})
			return
		}

		// 検索した本の情報をJSON形式でレスポンスとして返す
		c.JSON(200, book)

	})

	// 本情報を更新する。
	router.PATCH("/books/:id", func(c *gin.Context) {

		// 本の新規登録で送られてくるリクエストを型定義
		type UpdateBookRequest struct {
			Title  string `json:"title" binding:"required"`
			Author string `json:"author" binding:"required"`
			UserId int    `json:"user_id" binding:"required"`
		}

		// 引数で値を受け取るように変数を定義
		var book UpdateBookRequest

		// bookに受け取った値を格納
		if err := c.ShouldBindJSON(&book); err != nil {
			c.JSON(400, gin.H{"error": err.Error(), "message": "Invalid Book ID"})
			return
		}

		// URLパラメータから本のIDを取得
		bookIDStr := c.Param("id")

		// 数値に変換
		bookID, err := strconv.Atoi(bookIDStr)

		// パラメータが不正な場合はエラーを返して終了
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error(), "message": "could not translation string->int"})
			return
		}

		// 指定されたIDの本をデータベースから検索、更新
		update_book, err := client.Book.
			UpdateOneID(bookID).
			SetTitle(book.Title).
			SetAuthor(book.Author).
			Save(context.Background())

		// エラーならエラーを返して終了
		if err != nil {
			c.JSON(404, gin.H{"error": err.Error(), "message": "Couldn't update"})
			return
		}

		// 本の情報をJSON形式でレスポンスとして返す
		c.JSON(200, update_book)
	})

	// 本を削除
	router.DELETE("/books/:id", func(c *gin.Context) {
		// URLパラメータから本のIDを取得する
		bookIDStr := c.Param("id")

		// 数値に変換
		bookID, err := strconv.Atoi(bookIDStr)

		// パラメータが不正な場合はエラーを返して終了
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid Book ID"})
			return
		}

		// 指定されたIDの本をデータベースから検索、削除
		err = client.Book.DeleteOneID(bookID).Exec(context.Background())

		// エラーが起きた場合はエラーをあえして終了
		if err != nil {
			c.JSON(404, gin.H{"error": "Failed to delete"})
			return
		}

		c.JSON(200, gin.H{"message": "Delete completed"})
	})

	// サーバー起動
	router.Run(":8080")
}
