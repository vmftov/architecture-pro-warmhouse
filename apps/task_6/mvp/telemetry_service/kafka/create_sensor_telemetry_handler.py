from flask import Flask
from db.models import db, SensorTelemetry
from utils.dict_utils import dict_get_int, dict_get_float, dict_get_datetime

# Возвращает функцию для обработки телеметрии датчиков
def create_sensor_telemetry_handler(app: Flask):
    def handle_sensor_telemetry(data: dict):
        sensor_id = dict_get_int(data, 'sensor_id')
        if sensor_id is None: 
            return
        
        timestamp = dict_get_datetime(data, 'timestamp')
        if timestamp is None: 
            return
 
        value = dict_get_float(data, 'value')
        if value is None: 
            return
 
        with app.app_context():
            sensor_tel = SensorTelemetry.query.get(sensor_id)
            if sensor_tel:
                sensor_tel.timestamp = timestamp
                sensor_tel.value = value
            else:
                sensor_tel = SensorTelemetry(
                    sensor_id = sensor_id,
                    timestamp = timestamp,
                    value = value
                )
                db.session.add(sensor_tel)
            db.session.commit()

        print(f"Значение датчика {sensor_id} обновлено на ({timestamp}, {value})")

    return handle_sensor_telemetry

