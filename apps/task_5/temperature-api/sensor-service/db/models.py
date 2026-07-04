from flask_sqlalchemy import SQLAlchemy
from datetime import datetime, timezone

ILLEGAL_TEMP = 1_000_000

db = SQLAlchemy()

class Sensor(db.Model):
    __tablename__ = 'sensors'
    
    id = db.Column(db.Integer, primary_key=True)
    name = db.Column(db.String(100), nullable=False)
    type = db.Column(db.String(50), nullable=False)
    location = db.Column(db.String(100), nullable=False)
    value = db.Column(db.Float, default=ILLEGAL_TEMP)
    unit = db.Column(db.String(20))
    status = db.Column(db.String(20), nullable=False, default='inactive')
    last_updated = db.Column(db.DateTime(timezone=True), nullable=False, default=lambda: datetime.now(timezone.utc))
    created_at = db.Column(db.DateTime(timezone=True), nullable=False, default=lambda: datetime.now(timezone.utc))

    def to_dict(self):
        return {
            'id': self.id,
            'name': self.name,
            'type': self.type,
            'location': self.location,
            'value': self.value,
            'unit': self.unit,
            'status': self.status,
            'last_updated': self.last_updated.isoformat() if self.last_updated else None,
            'created_at': self.created_at.isoformat() if self.created_at else None
        }