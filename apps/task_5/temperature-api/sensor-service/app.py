
from flask import Flask, request, jsonify
from db.models import db, Sensor, ILLEGAL_TEMP
from utils.db_utils import init_db
from datetime import datetime, timezone
import random
import os

app = Flask(__name__)
app.config['SQLALCHEMY_DATABASE_URI'] = os.getenv('DATABASE_URL', 'postgresql://postgres:password@db:5432/smarthome')
db.init_app(app)


@app.route('/health', methods=['GET'])
def health_check():
    return jsonify({'status': 'ready'}), 200


# из задания не ясно какой именно путь имелся в виду
@app.route('/api/v1/sensors/temperature/<string:location>/', methods=['GET'])
@app.route('/temperature', methods=['GET'])
@app.route('/api/v1/temperature', methods=['GET'])
@app.route('/api/v1/sensors/temperature', methods=['GET'])
def get_temperature(location = None):
    if location is None:
        location = request.args.get('location')

    return jsonify({'location': location, 'temperature': get_random_temperature(), 'unit': '°C'}), 200


@app.route('/api/v1/sensors', methods=['GET'])
def get_all_sensors():
    result = []
    for sensor in Sensor.query.all():
        sensor_dict = sensor.to_dict()
        if sensor_dict['value'] == ILLEGAL_TEMP:
            sensor_dict['value'] = get_random_temperature()
        result.append(sensor_dict)

    return jsonify(result), 200


@app.route('/api/v1/sensors/<int:sensor_id>', methods=['GET'])
def get_sensor(sensor_id):
    sensor = Sensor.query.get(sensor_id)
    if not sensor:
        return jsonify(get_sensor_not_found_error()), 404
    
    sensor_dict = sensor.to_dict()
    if sensor_dict['value'] == ILLEGAL_TEMP:
        sensor_dict['value'] = get_random_temperature()

    return jsonify(sensor_dict), 200


@app.route('/api/v1/sensors', methods=['POST'])
def create_sensor():
    req_data = request.get_json()
    if not validate_sensor_required_fields(req_data):
        return jsonify(get_sensor_required_fields_missing_error()), 400
    
    utc_now = get_utc_now()
    sensor = Sensor(
        name = req_data['name'],
        type = req_data['type'],
        location = req_data['location'],
        unit = req_data.get('unit'),
        created_at = utc_now,
        last_updated = utc_now,
        value = ILLEGAL_TEMP,
        status = 'active'
    )
    db.session.add(sensor)
    db.session.commit()
    
    return jsonify({'message': 'Датчик успешно зарегистрирован', 'id': sensor.id}), 201


@app.route('/api/v1/sensors/<int:sensor_id>', methods=['PUT'])
def update_sensor(sensor_id):
    req_data = request.get_json()
    if not validate_sensor_required_fields(req_data):
        return jsonify(get_sensor_required_fields_missing_error()), 400
    
    sensor = Sensor.query.get(sensor_id)
    if not sensor:
        return jsonify(get_sensor_not_found_error()), 404
    
    sensor.name = req_data.get('name', sensor.name)
    sensor.type = req_data.get('type', sensor.type)
    sensor.location = req_data.get('location', sensor.location)
    sensor.unit = req_data.get('unit', sensor.unit)
    sensor.last_updated = get_utc_now()
    db.session.commit()

    return jsonify({'message': 'Датчик успешно обновлён'}), 200


@app.route('/api/v1/sensors/<int:sensor_id>', methods=['DELETE'])
def delete_sensor(sensor_id):
    sensor = Sensor.query.get(sensor_id)
    if sensor:
        db.session.delete(sensor)
        db.session.commit()

    return jsonify({'message': 'Датчик был успешно удалён'}), 200


@app.route('/api/v1/sensors/<int:sensor_id>/value', methods=['PATCH'])
def update_sensor_value(sensor_id):
    req_data = request.get_json()
    if ('status' in req_data) and (req_data['status'] not in ('active', 'inactive')):
        return jsonify({'message': 'Недопустимое значение статуса'}), 400

    sensor = Sensor.query.get(sensor_id)
    if not sensor:
        return jsonify(get_sensor_not_found_error()), 404
    
    if 'value' in req_data:
        sensor.value = req_data['value']
    if 'status' in req_data:
        sensor.status = req_data['status']
    
    sensor.last_updated = get_utc_now()
    db.session.commit()
    
    return jsonify({'message': 'Датчик успешно обновлён'}), 200


def validate_sensor_required_fields(req_data):
    return ('name' in req_data) and ('type' in req_data) and ('location' in req_data) and ('unit' in req_data)

def get_sensor_not_found_error():
    return {'message': 'Датчик не найден'}

def get_sensor_required_fields_missing_error():
    return {'message': 'Данные датчика должны содержать имя, тип, расположение и единицу измерения'}

def get_utc_now():
    return datetime.now(timezone.utc)

def get_random_temperature():
    return round(random.uniform(-42, 42), 1)


if __name__ == '__main__':
    with app.app_context():
        init_db()
    app.run(host='0.0.0.0', port=8081)