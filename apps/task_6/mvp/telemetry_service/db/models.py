from flask_sqlalchemy import SQLAlchemy
from datetime import datetime, timezone

db = SQLAlchemy()

class SensorTelemetry(db.Model):
    sensor_id = db.Column(db.Integer, primary_key=True)
    timestamp = db.Column(db.DateTime(timezone=True), nullable=False, default=lambda: datetime.now(timezone.utc))
    value = db.Column(db.Float, nullable=False, default=0)

    def to_dict(self):
        return {
            'sensor_id': self.sensor_id,
            'timestamp': self.timestamp.isoformat() if self.timestamp else None,
            'value': self.value
        }

class RelayValues(db.Model):
    relay_id = db.Column(db.Integer, primary_key=True)
    timestamp = db.Column(db.DateTime(timezone=True), nullable=False, default=lambda: datetime.now(timezone.utc))
    value = db.Column(db.Float, nullable=False, default=0)

    def to_dict(self):
        return {
            'relay_id': self.relay_id,
            'timestamp': self.timestamp.isoformat() if self.timestamp else None,
            'value': self.value
        }