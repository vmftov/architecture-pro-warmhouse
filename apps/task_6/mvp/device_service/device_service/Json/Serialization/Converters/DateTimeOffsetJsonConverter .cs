using System.Globalization;
using System.Text.Json;
using System.Text.Json.Serialization;

namespace device_service.Json.Serialization.Converters;

public class DateTimeOffsetJsonConverter : JsonConverter<DateTimeOffset> {
    public override void Write(Utf8JsonWriter writer, DateTimeOffset value, JsonSerializerOptions options) {
        var trimValue = new DateTimeOffset(value.Year, value.Month, value.Day, value.Hour, value.Minute, value.Second, value.Offset);
        writer.WriteStringValue(trimValue.ToString("yyyy-MM-ddTHH:mm:ssK"));
    }

    public override DateTimeOffset Read(ref Utf8JsonReader reader, Type typeToConvert, JsonSerializerOptions options) {
        var dateString = reader.GetString();
        if (dateString == null) return default;

        return DateTimeOffset.ParseExact(dateString, "o", CultureInfo.InvariantCulture, DateTimeStyles.RoundtripKind);
    }
}