from flask import Flask
from db.models import db, RelayValues
from utils.dict_utils import dict_get_int, dict_get_float, dict_get_datetime

# Возвращает функцию для обработки изменения значений реле
def create_relay_value_change_handler(app: Flask):
    def handle_relay_value_change(data: dict):
        relay_id = dict_get_int(data, 'relay_id')
        if relay_id is None: 
            return
        
        timestamp = dict_get_datetime(data, 'timestamp')
        if timestamp is None: 
            return
 
        value = dict_get_float(data, 'value')
        if value is None: 
            return
 
        with app.app_context():
            relay_vals = RelayValues.query.get(relay_id)
            if relay_vals:
                relay_vals.timestamp = timestamp
                relay_vals.value = value
            else:
                relay_vals = RelayValues(
                    relay_id = relay_id,
                    timestamp = timestamp,
                    value = value
                )
                db.session.add(relay_vals)
            db.session.commit()

        print(f"Значение реле {relay_id} обновлено на ({timestamp}, {value})")

    return handle_relay_value_change

