package v1

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/VsemoGosi-Adventure/events/internal/delivery/http/middleware"
	"github.com/VsemoGosi-Adventure/events/internal/domain"
	"github.com/VsemoGosi-Adventure/events/internal/dto"
)

type EventService interface {
	GetNearbyEvents(ctx context.Context, filter domain.GeoFilter) ([]dto.GeoEventItemResponse, error)
	GetEventByID(ctx context.Context, eventID uuid.UUID, userID *uuid.UUID) (*dto.EventDetailResponse, error)
	GetMyEvents(ctx context.Context, userID uuid.UUID, limit, offset int) ([]dto.GeoEventItemResponse, error)
	JoinEvent(ctx context.Context, eventID, userID uuid.UUID) error
	LeaveEvent(ctx context.Context, eventID, userID uuid.UUID) error
	GetParticipants(ctx context.Context, eventID, requesterID uuid.UUID) ([]dto.ParticipantResponse, error)
	CreatePing(ctx context.Context, eventID, authorID uuid.UUID, req dto.CreatePingRequest) (*dto.PingResponse, error)
	GetNearbyPings(ctx context.Context, lat, lng float64, limit, offset int) ([]dto.PingResponse, error)
	CheckAccess(ctx context.Context, eventID, userID uuid.UUID) (bool, error)
}

type EventHandler struct {
	service EventService
}

func NewEventHandler(service EventService) *EventHandler {
	return &EventHandler{
		service: service,
	}
}

func (h *EventHandler) RegisterRoutes(router *gin.Engine) {
	// TODO Повесить мидвари как будет готов сервис авторизации и траефик

	api := router.Group("/api/events")
	api.Use(middleware.RequestID(), middleware.SlogLogger())
	{
		api.GET("/nearby", h.GetNearby)
		api.GET("/:id", middleware.AuthOptional(), h.GetByID)
		api.GET("/pings/nearby", h.GetNearbyPings)
		api.GET("/:id/participants", h.GetParticipants)

		protected := api.Group("")
		protected.Use(middleware.AuthRequired())
		{
			protected.GET("/my", h.GetMyEvents)
			protected.POST("/:id/join", h.JoinEvent)
			protected.POST("/:id/leave", h.LeaveEvent)
			protected.POST("/:id/ping", h.CreatePing)
		}
	}

	internal := router.Group("/internal/v1/events")
	internal.Use(middleware.RequestID(), middleware.SlogLogger())
	{
		internal.GET("/:id/access", h.CheckAccess)
	}
}

// GetNearby godoc
// @Summary      Получение гео-ленты событий
// @Description  Возвращает список активных событий вокруг указанных координат с учетом радиуса.
// @Tags         events
// @Produce      json
// @Param        lat      query     number  true   "Широта (Latitude)"
// @Param        lng      query     number  true   "Долгота (Longitude)"
// @Param        radius   query     int     false  "Радиус в метрах (по умолчанию 3000)" default(3000)
// @Param        limit    query     int     false  "Лимит выборки (по умолчанию 20)" default(20)
// @Param        offset   query     int     false  "Смещение для пагинации (по умолчанию 0)" default(0)
// @Success      200      {array}   dto.GeoEventItemResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Failure      500      {object}  dto.ErrorResponse
// @Router       /api/events/nearby [get]
func (h *EventHandler) GetNearby(c *gin.Context) {
	lat, lng, ok := parseCoords(c)
	if !ok {
		return
	}

	radius, _ := strconv.Atoi(c.DefaultQuery("radius", "3000"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	filter := domain.GeoFilter{
		UserLat:      lat,
		UserLng:      lng,
		RadiusMeters: radius,
		Limit:        limit,
		Offset:       offset,
	}

	items, err := h.service.GetNearbyEvents(c.Request.Context(), filter)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, items)
}

// GetByID godoc
// @Summary      Детальная информация о встрече
// @Description  Возвращает полную карточку события по его ID. Если передан заголовок X-User-Id, проверяет статус участия.
// @Tags         events
// @Produce      json
// @Param        id       path      string  true   "UUID события"
// @Success      200      {object}  dto.EventDetailResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Failure      404      {object}  dto.ErrorResponse
// @Failure      500      {object}  dto.ErrorResponse
// @Router       /api/events/{id} [get]
func (h *EventHandler) GetByID(c *gin.Context) {
	eventID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var userIDPtr *uuid.UUID
	if uid, exists := middleware.GetUserID(c); exists {
		userIDPtr = &uid
	}

	event, err := h.service.GetEventByID(c.Request.Context(), eventID, userIDPtr)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, event)
}

// GetMyEvents godoc
// @Summary      Мои активные встречи
// @Description  Возвращает список встреч, в которых текущий пользователь принимает участие.
// @Tags         events
// @Produce      json
// @Param        limit    query     int     false  "Лимит выборки" default(20)
// @Param        offset   query     int     false  "Смещение" default(0)
// @Success      200      {array}   dto.GeoEventItemResponse
// @Failure      401      {object}  dto.ErrorResponse
// @Failure      500      {object}  dto.ErrorResponse
// @Router       /api/events/my [get]
func (h *EventHandler) GetMyEvents(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	items, err := h.service.GetMyEvents(c.Request.Context(), userID, limit, offset)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, items)
}

// JoinEvent godoc
// @Summary      Вступление во встречу
// @Description  Добавляет текущего пользователя в состав участников встречи в один клик.
// @Tags         events
// @Produce      json
// @Param        id       path      string  true   "UUID события"
// @Success      200      {object}  map[string]string
// @Failure      400      {object}  dto.ErrorResponse
// @Failure      401      {object}  dto.ErrorResponse
// @Failure      404      {object}  dto.ErrorResponse
// @Failure      409      {object}  dto.ErrorResponse "Конфликт: встреча заполнена или отменена"
// @Failure      500      {object}  dto.ErrorResponse
// @Router       /api/events/{id}/join [post]
func (h *EventHandler) JoinEvent(c *gin.Context) {
	eventID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	userID, _ := middleware.GetUserID(c)

	if err := h.service.JoinEvent(c.Request.Context(), eventID, userID); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "joined"})
}

// LeaveEvent godoc
// @Summary      Выход из встречи
// @Description  Переводит статус участника встречи в 'left', освобождая слот.
// @Tags         events
// @Produce      json
// @Param        id       path      string  true   "UUID события"
// @Success      200      {object}  map[string]string
// @Failure      400      {object}  dto.ErrorResponse
// @Failure      401      {object}  dto.ErrorResponse
// @Failure      500      {object}  dto.ErrorResponse
// @Router       /api/events/{id}/leave [post]
func (h *EventHandler) LeaveEvent(c *gin.Context) {
	eventID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	userID, _ := middleware.GetUserID(c)

	if err := h.service.LeaveEvent(c.Request.Context(), eventID, userID); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "left"})
}

// GetParticipants godoc
// @Summary      Список участников встречи
// @Description  Возвращает список пользователей, присоединившихся к событию (доступно только участникам).
// @Tags         events
// @Produce      json
// @Param        id       path      string  true   "UUID события"
// @Success      200      {array}   dto.ParticipantResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Failure      401      {object}  dto.ErrorResponse
// @Failure      403      {object}  dto.ErrorResponse "Доступ запрещен: пользователь не участник"
// @Failure      500      {object}  dto.ErrorResponse
// @Router       /api/events/{id}/participants [get]
func (h *EventHandler) GetParticipants(c *gin.Context) {
	eventID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	userID, _ := middleware.GetUserID(c)

	participants, err := h.service.GetParticipants(c.Request.Context(), eventID, userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, participants)
}

// CreatePing godoc
// @Summary      Отправка сигнала «Пинг рядом»
// @Description  Позволяет активному участнику встречи отправить пространственно-временной сигнал людям поблизости.
// @Tags         pings
// @Accept       json
// @Produce      json
// @Param        id       path      string                true   "UUID события"
// @Param        input    body      dto.CreatePingRequest true   "Параметры пинга"
// @Success      201      {object}  dto.PingResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Failure      401      {object}  dto.ErrorResponse
// @Failure      403      {object}  dto.ErrorResponse "Пользователь не является участником"
// @Failure      500      {object}  dto.ErrorResponse
// @Router       /api/events/{id}/ping [post]
func (h *EventHandler) CreatePing(c *gin.Context) {
	eventID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	userID, _ := middleware.GetUserID(c)

	var req dto.CreatePingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request body: " + err.Error()})
		return
	}

	ping, err := h.service.CreatePing(c.Request.Context(), eventID, userID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, ping)
}

// GetNearbyPings godoc
// @Summary      Активные пинги вокруг меня
// @Description  Возвращает список живых сигналов «Пинг», в радиус действия которых попадают переданные координаты.
// @Tags         pings
// @Produce      json
// @Param        lat      query     number  true   "Широта (Latitude)"
// @Param        lng      query     number  true   "Долгота (Longitude)"
// @Param        limit    query     int     false  "Лимит" default(20)
// @Param        offset   query     int     false  "Смещение" default(0)
// @Success      200      {array}   dto.PingResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Failure      500      {object}  dto.ErrorResponse
// @Router       /api/events/pings/nearby [get]
func (h *EventHandler) GetNearbyPings(c *gin.Context) {
	lat, lng, ok := parseCoords(c)
	if !ok {
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	pings, err := h.service.GetNearbyPings(c.Request.Context(), lat, lng, limit, offset)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, pings)
}

// CheckAccess godoc
// @Summary      Внутренняя проверка доступа к чату
// @Description  Используется микросервисом чата для проверки, имеет ли пользователь права на подключение к WS-комнате.
// @Tags         internal
// @Produce      json
// @Param        id       path      string  true   "UUID события"
// @Param        user_id  query     string  true   "UUID пользователя"
// @Success      200      {object}  dto.AccessCheckResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Failure      500      {object}  dto.ErrorResponse
// @Router       /internal/v1/events/{id}/access [get]
func (h *EventHandler) CheckAccess(c *gin.Context) {
	eventID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}

	userIDStr := c.Query("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid or missing 'user_id' query param"})
		return
	}

	hasAccess, err := h.service.CheckAccess(c.Request.Context(), eventID, userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.AccessCheckResponse{
		EventID:   eventID,
		UserID:    userID,
		HasAccess: hasAccess,
	})
}

func (h *EventHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrEventNotFound):
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrInvalidRadius):
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrNotParticipant):
		c.JSON(http.StatusForbidden, dto.ErrorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrEventFull),
		errors.Is(err, domain.ErrEventCancelled),
		errors.Is(err, domain.ErrEventFinished),
		errors.Is(err, domain.ErrAlreadyParticipant),
		errors.Is(err, domain.ErrPingExpired):
		c.JSON(http.StatusConflict, dto.ErrorResponse{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
	}
}

func parseUUIDParam(c *gin.Context, param string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(param))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid uuid in path parameter '" + param + "'"})
		return uuid.Nil, false
	}
	return id, true
}

func parseCoords(c *gin.Context) (float64, float64, bool) {
	latStr := c.Query("lat")
	lngStr := c.Query("lng")
	if latStr == "" || lngStr == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "query parameters 'lat' and 'lng' are required"})
		return 0, 0, false
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid 'lat' coordinate format"})
		return 0, 0, false
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid 'lng' coordinate format"})
		return 0, 0, false
	}

	return lat, lng, true
}
