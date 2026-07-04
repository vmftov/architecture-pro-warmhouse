using device_service.Json.Serialization.Converters;
using System.Text.Json;

namespace device_service.Json.Serialization;

public static class JsonSerializationUtils {
    static JsonSerializationUtils() {
        DefaultOptions = new JsonSerializerOptions();
        ConfigureOptions(DefaultOptions);
    }

    public static JsonSerializerOptions DefaultOptions { get; }

    public static void ConfigureOptions(JsonSerializerOptions options) {
        options.PropertyNamingPolicy = JsonNamingPolicy.SnakeCaseLower;
        options.Converters.Add(new DateTimeOffsetJsonConverter());
    }
}