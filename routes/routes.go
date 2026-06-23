package routes

import (
	"finance-tracker/controllers"
	"finance-tracker/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {

	r.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Finance Tracker API</title>
  <style>
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
      background: #0f0f0f;
      color: #fff;
      display: flex;
      align-items: center;
      justify-content: center;
      min-height: 100vh;
    }
    .card {
      text-align: center;
      padding: 48px;
      border: 1px solid #222;
      border-radius: 16px;
      background: #1a1a1a;
      max-width: 480px;
      width: 90%;
    }
    .badge {
      display: inline-block;
      background: #00c97720;
      color: #00c977;
      border: 1px solid #00c97740;
      border-radius: 999px;
      font-size: 13px;
      padding: 4px 14px;
      margin-bottom: 24px;
    }
    h1 { font-size: 28px; font-weight: 700; margin-bottom: 12px; }
    p  { color: #888; font-size: 15px; line-height: 1.6; margin-bottom: 28px; }
    .routes { text-align: left; background: #111; border-radius: 10px; padding: 16px 20px; }
    .route { display: flex; align-items: center; gap: 10px; padding: 6px 0; font-size: 14px; }
    .route:not(:last-child) { border-bottom: 1px solid #222; }
    .method {
      font-size: 11px; font-weight: 700; padding: 2px 8px;
      border-radius: 4px; min-width: 48px; text-align: center;
    }
    .post { background: #f59e0b20; color: #f59e0b; }
    .get  { background: #3b82f620; color: #60a5fa; }
    .path { color: #ccc; font-family: monospace; }
    .lock { color: #666; font-size: 12px; margin-left: auto; }
  </style>
</head>
<body>
  <div class="card">
    <div class="badge">✓ Successfully Hosted</div>
    <h1>Finance Tracker API</h1>
    <p>Go · Gin · PostgreSQL · JWT</p>
    <div class="routes">
      <div class="route">
        <span class="method post">POST</span>
        <span class="path">/register</span>
      </div>
      <div class="route">
        <span class="method post">POST</span>
        <span class="path">/login</span>
      </div>
      <div class="route">
        <span class="method post">POST</span>
        <span class="path">/api/transactions</span>
        <span class="lock">🔒 JWT</span>
      </div>
      <div class="route">
        <span class="method get">GET</span>
        <span class="path">/api/transactions</span>
        <span class="lock">🔒 JWT</span>
      </div>
      <div class="route">
        <span class="method get">GET</span>
        <span class="path">/api/dashboard</span>
        <span class="lock">🔒 JWT</span>
      </div>
    </div>
  </div>
</body>
</html>
`))
	})

	r.POST("/register",
		controllers.Register)

	r.POST("/login",
		controllers.Login)

	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())

	{
		api.POST(
			"/transactions",
			controllers.CreateTransaction,
		)

		api.GET(
			"/transactions",
			controllers.GetTransactions,
		)

		api.GET(
			"/dashboard",
			controllers.Dashboard,
		)
	}
}