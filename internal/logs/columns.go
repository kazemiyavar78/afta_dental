package logs

import "github.com/tpdenta/afta-reception/internal/platform/listquery"

// unknownActorLabel نام نمایشی ردیفی است که شناسه کاربر ندارد و نام فاعل هم ذخیره نشده.
// ردیف نسخه ۱ اغلب همین حالت را دارد. FAU_SAR.2: به‌جای شناسه خالی، برچسب خوانا نشان داده می‌شود.
const unknownActorLabel = "سیستم/ناشناس"

// actorUsernameSQL نام فاعل را از خود ردیف، و اگر خالی بود از کاربر، و اگر شناسه نداشت از برچسب ثابت می‌سازد.
// LEFT JOIN در کوئری لازم است. INNER JOIN ردیف بدون کاربر را حذف می‌کرد.
const actorUsernameSQL = "COALESCE(NULLIF(SecurityEvents.ActorUsername, N''), Users.Username, CASE WHEN SecurityEvents.UserID IS NULL THEN N'" + unknownActorLabel + "' END)"

// userJoin کاربران را برای نام ردیف‌های قدیمی وصل می‌کند و ردیف بدون کاربر را حذف نمی‌کند.
const userJoin = "LEFT JOIN Users ON Users.ID = SecurityEvents.UserID"

// eventSelectSQL ستون‌های پاسخ فهرست است. عبارت نام فاعل همان عبارت فیلتر است تا نتیجه فیلتر با نمایش یکی باشد.
const eventSelectSQL = "" +
	"SecurityEvents.ID AS id, " +
	"SecurityEvents.CreatedAt AS created_at, " +
	actorUsernameSQL + " AS actor_username, " +
	"SecurityEvents.TargetName AS target_name, " +
	"SecurityEvents.TargetType AS target_type, " +
	"SecurityEvents.EventType AS event_type, " +
	"SecurityEvents.Result AS result, " +
	"SecurityEvents.Reason AS reason, " +
	"SecurityEvents.Ip AS ip, " +
	"SecurityEvents.UserAgent AS user_agent, " +
	"SecurityEvents.Description AS description, " +
	"SecurityEvents.Details AS details"

// logSpec ستون‌های GET /api/logs است. همه قابل فیلتر و مرتب‌سازی‌اند.
// FAU_SAR.3 / R1-01: جستجو و مرتب‌سازی سمت سرور و فقط روی همین ستون‌هاست.
// D-16: IP متنی است و فاصله کنارش قبل از مقایسه حذف می‌شود.
func logSpec() listquery.Spec {
	return listquery.Spec{
		DefaultSort: "id",
		DefaultDesc: true,
		Columns: map[string]listquery.Column{
			"id":             {Expr: "SecurityEvents.ID", Type: listquery.KindNumber, Sortable: true, Filterable: true},
			"created_at":     {Expr: "SecurityEvents.CreatedAt", Type: listquery.KindDate, Sortable: true, Filterable: true},
			"actor_username": {Expr: actorUsernameSQL, Type: listquery.KindText, Sortable: true, Filterable: true},
			"target_name":    {Expr: "SecurityEvents.TargetName", Type: listquery.KindText, Sortable: true, Filterable: true},
			"target_type":    {Expr: "SecurityEvents.TargetType", Type: listquery.KindEnum, Sortable: true, Filterable: true},
			"event_type":     {Expr: "SecurityEvents.EventType", Type: listquery.KindEnum, Sortable: true, Filterable: true},
			"result":         {Expr: "SecurityEvents.Result", Type: listquery.KindEnum, Sortable: true, Filterable: true},
			"reason":         {Expr: "SecurityEvents.Reason", Type: listquery.KindEnum, Sortable: true, Filterable: true},
			"ip":             {Expr: "SecurityEvents.Ip", Type: listquery.KindText, Sortable: true, Filterable: true, Trim: true},
			"user_agent":     {Expr: "SecurityEvents.UserAgent", Type: listquery.KindText, Sortable: true, Filterable: true},
			"description":    {Expr: "SecurityEvents.Description", Type: listquery.KindText, Sortable: true, Filterable: true},
		},
	}
}
