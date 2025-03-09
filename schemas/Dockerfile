FROM eclipse-temurin:11-jre

RUN apt-get update && apt-get install -y wget jq curl

# RUN wget https://downloads.apache.org/avro/avro-1.11.2/java/avro-tools-1.11.2.jar -O /avro-tools.jar

COPY . .

RUN chmod +x /bash/producer_schema.sh

CMD ["sh", "/bash/producer_schema.sh"]