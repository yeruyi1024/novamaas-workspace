package router

import (
	"github.com/QuantumNous/new-api/controller"
	"github.com/QuantumNous/new-api/middleware"

	"github.com/gin-gonic/gin"
)

func SetVideoRouter(router *gin.Engine) {
	// Video proxy: accepts either session auth (dashboard) or token auth (API clients)
	videoProxyRouter := router.Group("/v1")
	videoProxyRouter.Use(middleware.RouteTag("relay"))
	videoProxyRouter.Use(middleware.TokenOrUserAuth())
	{
		videoProxyRouter.GET("/videos/:task_id/content-info", controller.GetVideoContentInfo)
		videoProxyRouter.GET("/videos/:task_id/content", controller.VideoProxy)
	}

	videoV1WriteRouter := router.Group("/v1")
	videoV1WriteRouter.Use(middleware.RouteTag("relay"))
	videoV1WriteRouter.Use(middleware.TokenAuth(), middleware.Distribute())
	{
		videoV1WriteRouter.POST("/video/generations", controller.RelayTask)
		videoV1WriteRouter.POST("/videos/:video_id/remix", controller.RelayTask)
		videoV1WriteRouter.POST("/videos", controller.RelayTask)
	}

	// Task detail reads are also used by the authenticated dashboard. The
	// response builder still enforces owner access, with an administrator-only
	// cross-user lookup for the task log page.
	videoV1ReadRouter := router.Group("/v1")
	videoV1ReadRouter.Use(middleware.RouteTag("relay"))
	videoV1ReadRouter.Use(middleware.TokenOrUserAuth(), middleware.Distribute())
	{
		videoV1ReadRouter.GET("/video/generations/:task_id", controller.RelayTaskFetch)
		videoV1ReadRouter.GET("/videos/:task_id", controller.RelayTaskFetch)
	}

	// Fire Ark native async content-generation API. These routes deliberately
	// stay outside /v1 so only ChannelTypeVolcNative can serve them.
	volcNativeWriteRouter := router.Group("/api/v3")
	volcNativeWriteRouter.Use(middleware.RouteTag("relay"))
	volcNativeWriteRouter.Use(middleware.SystemPerformanceCheck())
	volcNativeWriteRouter.Use(middleware.TokenAuth())
	volcNativeWriteRouter.Use(middleware.ModelRequestRateLimit())
	{
		volcNativeWriteRouter.POST("/contents/generations/tasks", middleware.Distribute(), controller.RelayTask)
		volcNativeWriteRouter.DELETE("/contents/generations/tasks/:task_id", controller.RelayVolcNativeTaskDelete)
	}

	volcNativeReadRouter := router.Group("/api/v3")
	volcNativeReadRouter.Use(middleware.RouteTag("relay"))
	volcNativeReadRouter.Use(middleware.SystemPerformanceCheck())
	volcNativeReadRouter.Use(middleware.TokenOrUserAuth())
	{
		volcNativeReadRouter.GET("/contents/generations/tasks", controller.RelayVolcNativeTaskList)
		volcNativeReadRouter.GET("/contents/generations/tasks/:task_id", controller.RelayVolcNativeTaskFetch)
	}

	klingV1Router := router.Group("/kling/v1")
	klingV1Router.Use(middleware.RouteTag("relay"))
	klingV1Router.Use(middleware.KlingRequestConvert(), middleware.TokenAuth(), middleware.Distribute())
	{
		klingV1Router.POST("/videos/text2video", controller.RelayTask)
		klingV1Router.POST("/videos/image2video", controller.RelayTask)
		klingV1Router.GET("/videos/text2video/:task_id", controller.RelayTaskFetch)
		klingV1Router.GET("/videos/image2video/:task_id", controller.RelayTaskFetch)
	}

	// Jimeng official API routes - direct mapping to official API format
	jimengOfficialGroup := router.Group("jimeng")
	jimengOfficialGroup.Use(middleware.RouteTag("relay"))
	jimengOfficialGroup.Use(middleware.JimengRequestConvert(), middleware.TokenAuth(), middleware.Distribute())
	{
		// Maps to: /?Action=CVSync2AsyncSubmitTask&Version=2022-08-31 and /?Action=CVSync2AsyncGetResult&Version=2022-08-31
		jimengOfficialGroup.POST("/", controller.RelayTask)
	}
}
