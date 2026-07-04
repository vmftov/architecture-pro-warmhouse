using System.ComponentModel.DataAnnotations.Schema;

namespace device_service.Models;

public class Sensor {
    public int SensorId { get; set; }

    public string Name { get; set; }

    public string Type { get; set; }

    public string Location { get; set; }

    public string? Unit { get; set; }

    public string? Status { get; set; }

    public DateTimeOffset CreatedAt { get; set; }

    public DateTimeOffset LastUpdated { get; set; }
}

public class Relay {
    public int RelayId { get; set; }

    public string Name { get; set; }

    public string Type { get; set; }

    public string Location { get; set; }

    public string? Unit { get; set; }

    public string? Status { get; set; }

    public double DesiredValue { get; set; }

    public DateTimeOffset CreatedAt { get; set; }

    public DateTimeOffset LastUpdated { get; set; }
}