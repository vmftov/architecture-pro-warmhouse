from confluent_kafka import Consumer, KafkaError, KafkaException
import json
import threading
import time

class KafkaConsumer:
    def __init__(self, bootstrap_servers, topic, group_id, message_handler):
        self.bootstrap_servers = bootstrap_servers
        self.topic = topic
        self.group_id = group_id
        self.message_handler = message_handler

        self.lock = threading.Lock()
        self.consumer = None
        self.thread = None
        self.is_running = False

        print("KafkaConsumerService инициализирован")


    def start_async(self):
        try:
            started = False

            with self.lock:
                if self.is_running or (self.thread and self.thread.is_alive()):
                    return
                
                self.consumer = self._create_consumer()
                self.is_running = True

                self.thread = threading.Thread(target = self._thread_loop, daemon = True)
                self.thread.start()

                started = True

            if started:
                print("KafkaConsumerService запущен")
        except Exception as exc:
            print(f"При запуске KafkaConsumerService произошла ошибка: {exc}")
            raise
    

    def stop(self):
        try:
            stopped = False
        
            with self.lock:
                if not self.is_running:
                    return
                
                self.is_running = False

                if self.consumer:
                    self.consumer.close()
                    self.consumer = None

                stopped = True

            if stopped:
                print("KafkaConsumerService остановлен")
        except Exception as exc:
            print(f"При остановке KafkaConsumerService произошла ошибка: {exc}")
            raise
    

    def _create_consumer(self):
        config = {
            'bootstrap.servers': self.bootstrap_servers,
            'group.id': self.group_id,
            'auto.offset.reset': 'earliest',
            'enable.auto.commit': True,
            'allow.auto.create.topics': True
        }
        
        consumer = Consumer(config)
        consumer.subscribe([self.topic])
        
        return consumer
    

    def _thread_loop(self):
        try:
            while True:
                is_running = False
                msg = None

                with self.lock:
                    is_running = self.is_running
                    if self.is_running: 
                        msg = self.consumer.poll(1.0)

                if not is_running:
                    break
                if msg is None:
                    continue

                err = msg.error()
                if err:
                    err_code = err.code()
                    if err_code == KafkaError._PARTITION_EOF:
                        continue
                    if err_code == KafkaError.UNKNOWN_TOPIC_OR_PART or err_code == KafkaError.UNKNOWN_TOPIC_ID:
                        time.sleep(1.0)
                        continue
                    raise KafkaException(f"Kafka вернул ошибку: {msg.error()}")
                
                self._process_message(msg)
        except KeyboardInterrupt:
            print("KafkaConsumerService остановлен KeyboardInterrupt")
        except Exception as exc:
            print(f"Во время работы KafkaConsumerService._thread_loop произошла ошибка: {exc}")
        finally:
            self.stop()
    

    def _process_message(self, msg):
        data: dict = None

        try:
            data = json.loads(msg.value().decode('utf-8'))
        except json.JSONDecodeError as exc:
            print(f"При парсинге сообщения произошла ошибка: {msg.value().decode('utf-8')}")
            return
        except Exception as exc:
            print(f"При чтении данных сообщения произошла ошибка. Сообщение: {data}. Ошибка: {exc}")
            return

        try:
            self.message_handler(data)
        except Exception as exc:
            print(f"При обработке сообщения произошла ошибка. Сообщение: {data}. Ошибка: {exc}")
