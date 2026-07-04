using device_service.Models;
using Microsoft.EntityFrameworkCore;

namespace device_service.Db;

public class DevicesDbContext : DbContext {
    public DevicesDbContext(DbContextOptions<DevicesDbContext> options) //
        : base(options) {
    }

    public DbSet<Sensor> Sensors { get; set; }

    public DbSet<Relay> Relays { get; set; }

    protected override void OnModelCreating(ModelBuilder modelBuilder) {
        modelBuilder.Entity<Sensor>(entity => {
            entity.ToTable("sensors");

            entity.HasKey(e => e.SensorId);
            entity.Property(e => e.SensorId).HasColumnName("sensor_id").UseIdentityColumn();

            entity.Property(e => e.Name).HasColumnName("name").IsRequired().HasMaxLength(255);
            entity.Property(e => e.Type).HasColumnName("type").IsRequired().HasMaxLength(64);
            entity.Property(e => e.Location).HasColumnName("location").IsRequired().HasMaxLength(255);
            entity.Property(e => e.Unit).HasColumnName("unit").HasMaxLength(16);
            entity.Property(e => e.Status).HasColumnName("status").HasMaxLength(16);
            entity.Property(e => e.CreatedAt).HasColumnName("created_at").IsRequired().HasDefaultValueSql("NOW()");
            entity.Property(e => e.LastUpdated).HasColumnName("last_updated").IsRequired().HasDefaultValueSql("NOW()");
        });

        modelBuilder.Entity<Relay>(entity => {
            entity.ToTable("relays");

            entity.HasKey(e => e.RelayId);
            entity.Property(e => e.RelayId).HasColumnName("relay_id").UseIdentityColumn();

            entity.Property(e => e.Name).HasColumnName("name").IsRequired().HasMaxLength(255);
            entity.Property(e => e.Type).HasColumnName("type").IsRequired().HasMaxLength(64);
            entity.Property(e => e.Location).HasColumnName("location").IsRequired().HasMaxLength(255);
            entity.Property(e => e.Unit).HasColumnName("unit").HasMaxLength(16);
            entity.Property(e => e.Status).HasColumnName("status").HasMaxLength(16);
            entity.Property(e => e.DesiredValue).HasColumnName("desired_value").IsRequired().HasDefaultValue(0.0);
            entity.Property(e => e.CreatedAt).HasColumnName("created_at").IsRequired().HasDefaultValueSql("NOW()");
            entity.Property(e => e.LastUpdated).HasColumnName("last_updated").IsRequired().HasDefaultValueSql("NOW()");
        });

        base.OnModelCreating(modelBuilder);
    }
}