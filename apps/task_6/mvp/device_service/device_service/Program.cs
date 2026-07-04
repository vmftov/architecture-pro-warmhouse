using device_service.Db;
using device_service.Json.Serialization;
using device_service.Kafka;
using device_service.Models;
using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;

// ReSharper disable AccessToDisposedClosure

#region Init

var builder = WebApplication.CreateBuilder(args);

builder.Services.AddControllers().AddJsonOptions(options => {
    JsonSerializationUtils.ConfigureOptions(options.JsonSerializerOptions);
});
builder.Services.ConfigureHttpJsonOptions(options => {
    JsonSerializationUtils.ConfigureOptions(options.SerializerOptions);
});
builder.Services.AddDbContext<DevicesDbContext>(options => {
    options.UseNpgsql(Environment.GetEnvironmentVariable("DATABASE_CON_STRING") ?? "Host=localhost;Database=devices_db;Username=postgres;Password=1234Qwer!");
});

builder.Services.AddEndpointsApiExplorer();
builder.Services.AddSwaggerGen();

var app = builder.Build();

if (app.Environment.IsDevelopment()) {
    app.UseHttpsRedirection();

    app.UseSwagger();
    app.UseSwaggerUI(c => {
        c.SwaggerEndpoint("/swagger/v1/swagger.json", "My API V1");
    });
}

using var sensorTelemetryProducer = new KafkaProducer(                                 //
    Environment.GetEnvironmentVariable("KAFKA_BOOTSTRAP_SERVERS") ?? "localhost:9092", //
    Environment.GetEnvironmentVariable("KAFKA_SENSOR_TELEMETRY_TOPIC") ?? "smarthome.sensor.telemetry");

using var relayValueChangedProducer = new KafkaProducer(                               //
    Environment.GetEnvironmentVariable("KAFKA_BOOTSTRAP_SERVERS") ?? "localhost:9092", //
    Environment.GetEnvironmentVariable("KAFKA_RELAY_VALUES_CHANGED_TOPIC") ?? "smarthome.relay.value-changed");

#endregion

#region General endpoints

app.MapGet("/health", () => Results.Ok(new { Status = "ready" }));

#endregion

#region Sensor endpoints

app.MapGet("/api/v1/sensors", async (string? location, DevicesDbContext db) => {
    if (string.IsNullOrEmpty(location)) {
        var sensors = await db.Sensors.ToListAsync();
        return Results.Ok(sensors);
    }
    var locationSensors = await db.Sensors.Where(s => s.Location == location).ToListAsync();
    if (locationSensors.Count > 0) {
        return Results.Ok(locationSensors);
    }
    var defaulSensor = await db.Sensors.FirstOrDefaultAsync();
    if (defaulSensor != null) {
        return Results.Ok(new[] { defaulSensor });
    }
    return Results.Ok(Array.Empty<Sensor>());
});

app.MapGet("/api/v1/sensors/{sensorId:int}", async (int sensorId, DevicesDbContext db) => {
    var sensor = await db.Sensors.FindAsync(sensorId);
    if (sensor is null) {
        return Results.NotFound(new { Message = "Датчик не найден" });
    }

    return Results.Ok(sensor);
});

app.MapPost("/api/v1/sensors", async ([FromBody] Sensor sensor, DevicesDbContext db) => {
    if (!ValidateSensorRequiredFields(sensor, out var errorMessage)) {
        return Results.BadRequest(new { Message = errorMessage });
    }

    sensor.SensorId = 0;
    sensor.Status = "active";
    sensor.CreatedAt = DateTimeOffset.UtcNow;
    sensor.LastUpdated = sensor.CreatedAt;
    db.Sensors.Add(sensor);
    await db.SaveChangesAsync();

    await sensorTelemetryProducer.ProduceAsync(new {
        SensorId = sensor.SensorId,
        Timestamp = sensor.CreatedAt,
        Value = 1_000_000
    });

    return Results.Created($"/api/v1/sensors/{sensor.SensorId}", new {
        Message = "Датчик успешно зарегистрирован",
        Id = sensor.SensorId
    });
});

app.MapPut("/api/v1/sensors/{sensorId:int}", async (int sensorId, [FromBody] Sensor updatedSensor, DevicesDbContext db) => {
    if (!ValidateSensorRequiredFields(updatedSensor, out var errorMessage)) {
        return Results.BadRequest(new { Message = errorMessage });
    }

    var sensor = await db.Sensors.FindAsync(sensorId);
    if (sensor is null) {
        return Results.NotFound(new { Message = "Датчик не найден" });
    }

    sensor.Name = updatedSensor.Name;
    sensor.Type = updatedSensor.Type;
    sensor.Location = updatedSensor.Location;
    sensor.Unit = updatedSensor.Unit;
    sensor.LastUpdated = DateTimeOffset.UtcNow;
    await db.SaveChangesAsync();

    return Results.Ok(new { Message = "Датчик успешно обновлен" });
});

app.MapDelete("/api/v1/sensors/{sensorId:int}", async (int sensorId, DevicesDbContext db) => {
    var sensor = await db.Sensors.FindAsync(sensorId);
    if (sensor is not null) {
        db.Sensors.Remove(sensor);
        await db.SaveChangesAsync();
    }

    return Results.Ok(new { Message = "Датчик был успешно удален" });
});

app.MapPatch("/api/v1/sensors/{sensorId:int}/value", async (int sensorId, [FromBody] Sensor data, DevicesDbContext db) => {
    if (data.Status is not null && data.Status != "active" && data.Status != "inactive") {
        return Results.BadRequest(new { Message = "Недопустимое значение статуса" });
    }

    var sensor = await db.Sensors.FindAsync(sensorId);
    if (sensor is null) {
        return Results.NotFound(new { Message = "Датчик не найден" });
    }

    if (!string.IsNullOrEmpty(data.Status)) {
        sensor.Status = data.Status;
    }
    sensor.LastUpdated = DateTimeOffset.UtcNow;
    await db.SaveChangesAsync();

    return Results.Ok(new { Message = "Датчик успешно обновлен" });
});

static bool ValidateSensorRequiredFields(Sensor sensor, out string errorMessage) {
    errorMessage = string.Empty;

    if (string.IsNullOrWhiteSpace(sensor.Name)) {
        errorMessage = "У датчика должны быть заполнены имя, тип, расположение и единица измерения";
    } else if (string.IsNullOrWhiteSpace(sensor.Type)) {
        errorMessage = "У датчика должны быть заполнены имя, тип, расположение и единица измерения";
    } else if (string.IsNullOrWhiteSpace(sensor.Location)) {
        errorMessage = "У датчика должны быть заполнены имя, тип, расположение и единица измерения";
    } else if (string.IsNullOrWhiteSpace(sensor.Unit)) {
        errorMessage = "У датчика должны быть заполнены имя, тип, расположение и единица измерения";
    }

    return string.IsNullOrEmpty(errorMessage);
}

#endregion

#region Relay endpoints

app.MapGet("/api/v1/relays", async (DevicesDbContext db) => {
    var relays = await db.Relays.ToListAsync();
    return Results.Ok(relays);
});

app.MapGet("/api/v1/relays/{relayId:int}", async (int relayId, DevicesDbContext db) => {
    var relay = await db.Relays.FindAsync(relayId);
    if (relay is null) {
        return Results.NotFound(new { Message = "Реле не найдено" });
    }

    return Results.Ok(relay);
});

app.MapPost("/api/v1/relays", async ([FromBody] Relay relay, DevicesDbContext db) => {
    if (!ValidateRelayRequiredFields(relay, out var errorMessage)) {
        return Results.BadRequest(new { Message = errorMessage });
    }

    relay.RelayId = 0;
    relay.CreatedAt = DateTimeOffset.UtcNow;
    relay.LastUpdated = relay.CreatedAt;
    relay.Status = "active";
    db.Relays.Add(relay);
    await db.SaveChangesAsync();

    await relayValueChangedProducer.ProduceAsync(new {
        RelayId = relay.RelayId,
        Timestamp = relay.CreatedAt,
        Value = relay.DesiredValue
    });

    return Results.Created($"/api/v1/relays/{relay.RelayId}", new { Message = "Реле успешно зарегистрировано", Id = relay.RelayId });
});

app.MapPut("/api/v1/relays/{relayId:int}", async (int relayId, [FromBody] Relay newRelay, DevicesDbContext db) => {
    if (!ValidateRelayRequiredFields(newRelay, out var errorMessage)) {
        return Results.BadRequest(new { Message = errorMessage });
    }

    var relay = await db.Relays.FindAsync(relayId);
    if (relay is null) {
        return Results.NotFound(new { Message = "Реле не найдено" });
    }

    relay.Name = newRelay.Name;
    relay.Type = newRelay.Type;
    relay.Location = newRelay.Location;
    relay.Unit = newRelay.Unit;
    relay.LastUpdated = DateTimeOffset.UtcNow;
    await db.SaveChangesAsync();

    return Results.Ok(new { Message = "Реле успешно обновлено" });
});

app.MapPatch("/api/v1/relays/{relayId:int}/values", async (int relayId, [FromBody] RelayValueChangeDto data, DevicesDbContext db) => {
    var relay = await db.Relays.FindAsync(relayId);
    if (relay is null) {
        return Results.NotFound(new { Message = "Реле не найдено" });
    }

    if (data.Value.HasValue) {
        relay.DesiredValue = data.Value.Value;
        relay.LastUpdated = DateTimeOffset.UtcNow;
        await db.SaveChangesAsync();
    }

    await relayValueChangedProducer.ProduceAsync(new {
        RelayId = relay.RelayId,
        Timestamp = relay.CreatedAt,
        Value = relay.DesiredValue
    });

    return Results.Ok(new { Message = "Реле успешно обновлено" });
});

app.MapDelete("/api/v1/relays/{relayId:int}", async (int relayId, DevicesDbContext db) => {
    var relay = await db.Relays.FindAsync(relayId);
    if (relay is not null) {
        db.Relays.Remove(relay);
        await db.SaveChangesAsync();
    }

    return Results.Ok(new { Message = "Реле было успешно удалено" });
});

static bool ValidateRelayRequiredFields(Relay relay, out string errorMessage) {
    errorMessage = string.Empty;

    if (string.IsNullOrWhiteSpace(relay.Name)) {
        errorMessage = "У реле должны быть заполнены имя, тип, расположение и единица измерения";
    } else if (string.IsNullOrWhiteSpace(relay.Type)) {
        errorMessage = "У реле должны быть заполнены имя, тип, расположение и единица измерения";
    } else if (string.IsNullOrWhiteSpace(relay.Location)) {
        errorMessage = "У реле должны быть заполнены имя, тип, расположение и единица измерения";
    } else if (string.IsNullOrWhiteSpace(relay.Unit)) {
        errorMessage = "У реле должны быть заполнены имя, тип, расположение и единица измерения";
    }

    return string.IsNullOrEmpty(errorMessage);
}

#endregion

app.Run();