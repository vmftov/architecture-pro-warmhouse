from flask import Flask, jsonify
from db.models import db, SensorTelemetry, RelayValues
from utils.db_utils import init_db
import os
import random
from kafka.kafka_consumer import KafkaConsumer
from kafka.create_sensor_telemetry_handler import create_sensor_telemetry_handler
from kafka.create_relay_value_change_handler import create_relay_value_change_handler


ILLEGAL_TEMP = 1_000_000

app = Flask(__name__)
app.config['SQLALCHEMY_DATABASE_URI'] = os.getenv('DATABASE_URL', 'postgresql://postgres:1234Qwer%21@localhost:5432/telemetry_db')
db.init_app(app)

kafka_sensor_consumer = KafkaConsumer(
    os.getenv('KAFKA_BOOTSTRAP_SERVERS', 'localhost:9092'), 
    os.getenv('KAFKA_SENSOR_TELEMETRY_TOPIC', 'smarthome.sensor.telemetry'), 
    'smarthome.sensor.telemetry.consumer',
    create_sensor_telemetry_handler(app)
)
kafka_relay_consumer = KafkaConsumer(
    os.getenv('KAFKA_BOOTSTRAP_SERVERS', 'localhost:9092'), 
    os.getenv('KAFKA_RELAY_VALUES_CHANGED_TOPIC', 'smarthome.relay.value-changed'), 
    'smarthome.relay.value-change.consumer',
    create_relay_value_change_handler(app)
)


#  general


@app.route('/health', methods=['GET'])
def health_check():
    return jsonify({'status': 'ready'}), 200


# sensors


@app.route('/api/v1/sensor-values', methods=['GET'])
def get_all_sensor_values():
    sensor_vals = []
    for sensor in SensorTelemetry.query.all():
        sensor_dict = sensor.to_dict()
        if sensor_dict['value'] == ILLEGAL_TEMP:
            sensor_dict['value'] = get_random_temperature()
        sensor_vals.append(sensor_dict)
    
    return jsonify(sensor_vals), 200


@app.route('/api/v1/sensor-values/<int:sensor_id>', methods=['GET'])
def get_sensor(sensor_id):
    sensor_val = SensorTelemetry.query.get(sensor_id)
    if not sensor_val:
        return jsonify(get_sensor_not_found_error()), 404
    
    sensor_dict = sensor_val.to_dict()
    if sensor_dict['value'] == ILLEGAL_TEMP:
        sensor_dict['value'] = get_random_temperature()

    return jsonify(sensor_dict), 200


@app.route('/api/v1/sensor-values/<int:sensor_id>', methods=['DELETE'])
def delete_sensor(sensor_id):
    sensor_val = SensorTelemetry.query.get(sensor_id)
    if sensor_val:
        db.session.delete(sensor_val)
        db.session.commit()
    return jsonify({'message': 'Телеметрия датчика была успешно удалена'}), 200


def get_sensor_not_found_error():
    return {'message': 'Датчик не найден'}


# relays


@app.route('/api/v1/relay-values', methods=['GET'])
def get_all_relay_values():
    relay_vals = []
    for relay in RelayValues.query.all():
        relay_vals.append(relay.to_dict())
    return jsonify(relay_vals), 200


@app.route('/api/v1/relay-values/<int:relay_id>', methods=['GET'])
def get_relay_value(relay_id):
    relay_val = RelayValues.query.get(relay_id)
    if not relay_val:
        return jsonify(get_relay_not_found_error()), 404
    return jsonify(relay_val.to_dict()), 200


@app.route('/api/v1/relay-values/<int:relay_id>', methods=['DELETE'])
def delete_relay(relay_id):
    relay_val = RelayValues.query.get(relay_id)
    if relay_val:
        db.session.delete(relay_val)
        db.session.commit()
        return jsonify({'message': 'Телеметрия реле была успешно удалена'}), 200
    return jsonify({'message': 'Реле не найдено'}), 404


def get_relay_not_found_error():
    return {'message': 'Реле не найдено'}


# utils


def get_random_temperature():
    return round(random.uniform(-42, 42), 1)


# startup


if __name__ == '__main__':
    with app.app_context():
        init_db()
    
    try:
        kafka_sensor_consumer.start_async()
        kafka_relay_consumer.start_async()
        app.run(host='0.0.0.0', port=5001)
    finally:
        kafka_sensor_consumer.stop()
        kafka_relay_consumer.stop()
