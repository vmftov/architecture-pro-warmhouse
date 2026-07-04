from datetime import datetime

def dict_get_int(data: dict, name: str):
    val = data.get(name)
    if val is None:
        return None
    
    if not isinstance(val, int):
        print(f"Поле {name} должно быть целым числом, получено {type(val)}")
        return None
    
    return val


def dict_get_float(data: dict, name: str):
    val = data.get(name)
    if val is None:
        print(f"Поле {name} отсутствует")
        return None
    
    if not isinstance(val, (int, float)):
        print(f"Поле {name} должно быть числом, получено {type(val)}")
        return None
    
    return float(val)


def dict_get_datetime(data: dict, name: str):
    val_str = data.get(name)
    if val_str is None:
        print(f"Поле {name} отсутствует")
        return None
    
    if not isinstance(val_str, str):
        print(f"Поле {name} должно быть строкой (датой), получено {type(val_str)}")
        return None
    
    val = None
    try:
        val = datetime.fromisoformat(val_str.replace('Z', '+00:00'))
    except Exception as exc:
        print(f"Поле {name} должно быть строкой (датой), получено '{val_str}'. Ошибка: {exc}")
        return None
    
    return val
