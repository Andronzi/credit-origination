SCHEMA_FILE="schemas/avro/credit/v1/StatusEvent.avsc"

if ! avro-tools compile schema "$SCHEMA_FILE" /tmp > /dev/null 2>&1; then
  echo "Ошибка: Схема $SCHEMA_FILE невалидна"
  avro-tools compile schema "$SCHEMA_FILE" /tmp
  exit 1
fi


SCHEMA=$(cat "$SCHEMA_FILE")
JSON_DATA=$(echo '{}' | jq --arg schema "$SCHEMA" '.schema = $schema')

curl -X POST -H "Content-Type: application/vnd.schemaregistry.v1+json" \
  --data "$JSON_DATA" \
  http://localhost:8081/subjects/application/versions