package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/DanielGregorini/notepad-on/controller"
)


func PageRoute(router *gin.Engine, pageController controller.PageController) {
	router.GET("/page/:slug", pageController.FetchAndUpdateText)
}
