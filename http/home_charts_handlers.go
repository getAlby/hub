package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

const defaultHomeChartsDays = 7

func (httpSvc *HttpService) homeChartsHandler(c echo.Context) error {
	from := uint64(time.Now().Add(-defaultHomeChartsDays * 24 * time.Hour).Unix())
	if fromRaw := c.QueryParam("from"); fromRaw != "" {
		parsed, err := strconv.ParseUint(fromRaw, 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Message: "invalid query parameter: from",
			})
		}
		from = parsed
	}

	response, err := httpSvc.api.GetHomeChartsData(c.Request().Context(), from)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response)
}
