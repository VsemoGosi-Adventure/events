# events
Микросервис событий и геопоиска

## Redpanda consumer

Сервис читает JSON-сообщения о найденных событиях из Redpanda и сохраняет их в Postgres. Настройки задаются в `internal/config/.env` (пример: `internal/config/.env.example`): `CONSUMER_BROKERS` (адреса через запятую), `CONSUMER_TOPIC` и `CONSUMER_GROUP_ID`.

Повторное сообщение с той же парой `external_source` и `external_source_id` обновляет событие и его локацию. Строковые теги сохраняются в `events.events.tags`. При ошибке базы consumer повторяет обработку; некорректный JSON или сообщение с недопустимыми полями пропускается с предупреждением в логе.
