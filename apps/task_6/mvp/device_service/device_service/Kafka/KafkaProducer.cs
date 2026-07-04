using Confluent.Kafka;
using System.Text.Json;
using device_service.Json.Serialization;

namespace device_service.Kafka;

public class KafkaProducer : IDisposable {
    private readonly string _topic;
    private readonly IProducer<Null, string> _producer;

    public KafkaProducer(string bootstrapServers, string topic) {
        _topic = topic;

        var config = new ProducerConfig {
            BootstrapServers = bootstrapServers,
            ClientId = "smarthome.device-service",
            MessageSendMaxRetries = 3
        };
        _producer = new ProducerBuilder<Null, string>(config).Build();
    }

    public async Task ProduceAsync<T>(T value) {
        try {
            var jsonMessage = JsonSerializer.Serialize(value, JsonSerializationUtils.DefaultOptions);

            var deliveryResult = await _producer.ProduceAsync(_topic, new Message<Null, string> {
                Value = jsonMessage
            });

            Console.WriteLine($"Сообщение отправлено в топик. Сообщение {jsonMessage}, Топик: {_topic}, Партиция: {deliveryResult.Partition}, Оффсет: {deliveryResult.Offset}");
        } catch (ProduceException<Null, string> exc) {
            Console.WriteLine($"Сообщение не было отправлено в топик из-за ошибки: {exc.Message}. " + exc);
        }
    }

    public void Dispose() {
        _producer.Dispose();
    }
}