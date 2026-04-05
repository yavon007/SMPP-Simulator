#!/usr/bin/env python3
"""
简单的 SMPP 客户端，用于端到端测试
支持 bind_transmitter, bind_receiver, bind_transceiver, submit_sm, unbind

用法:
    python3 smpp_client.py bind --host 127.0.0.1 --port 2775 --system-id test --password test --type transceiver
    python3 smpp_client.py send --host 127.0.0.1 --port 2775 --system-id test --password test --from 10086 --to 13800138000 --message "Hello"
    python3 smpp_client.py unbind --host 127.0.0.1 --port 2775
"""

import socket
import struct
import argparse
import sys

# SMPP Command IDs
CMD_BIND_TRANSMITTER = 0x00000002
CMD_BIND_RECEIVER = 0x00000001
CMD_BIND_TRANSCEIVER = 0x00000009
CMD_UNBIND = 0x00000006
CMD_SUBMIT_SM = 0x00000004
CMD_BIND_TRANSMITTER_RESP = 0x80000002
CMD_BIND_RECEIVER_RESP = 0x80000001
CMD_BIND_TRANSCEIVER_RESP = 0x80000009
CMD_UNBIND_RESP = 0x80000006
CMD_SUBMIT_SM_RESP = 0x80000004
CMD_GENERIC_NACK = 0x80000000

# SMPP Command Status
STATUS_OK = 0x00000000

class SMPPClient:
    def __init__(self, host, port):
        self.host = host
        self.port = port
        self.socket = None
        self.sequence = 1

    def connect(self):
        """连接到 SMPP 服务器"""
        self.socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.socket.settimeout(10)
        try:
            self.socket.connect((self.host, self.port))
            return True
        except Exception as e:
            print(f"连接失败: {e}")
            return False

    def close(self):
        """关闭连接"""
        if self.socket:
            self.socket.close()
            self.socket = None

    def _next_sequence(self):
        """获取下一个序列号"""
        seq = self.sequence
        self.sequence += 1
        return seq

    def _build_pdu(self, command_id, sequence_num, body=b''):
        """构建 PDU"""
        length = 16 + len(body)  # header (16 bytes) + body
        header = struct.pack('>IIII', length, command_id, STATUS_OK, sequence_num)
        return header + body

    def _parse_pdu(self, data):
        """解析 PDU"""
        if len(data) < 16:
            return None, None, None, None

        length, command_id, command_status, sequence_num = struct.unpack('>IIII', data[:16])
        body = data[16:length]
        return command_id, command_status, sequence_num, body

    def _encode_cstring(self, s):
        """编码 C 字符串 (null-terminated)"""
        return s.encode('ascii') + b'\x00'

    def _read_response(self, expected_command_id, expected_sequence):
        """读取响应"""
        try:
            # 先读取 header
            header = self.socket.recv(16)
            if len(header) < 16:
                return None, "接收响应头失败"

            length, command_id, command_status, sequence_num = struct.unpack('>IIII', header)

            # 读取剩余数据
            remaining = length - 16
            body = b''
            while len(body) < remaining:
                chunk = self.socket.recv(remaining - len(body))
                if not chunk:
                    break
                body += chunk

            if command_id == CMD_GENERIC_NACK:
                return None, f"服务器返回 NACK, status: 0x{command_status:08x}"

            if command_id != expected_command_id:
                return None, f"意外的响应类型: 0x{command_id:08x}"

            if command_status != STATUS_OK:
                return None, f"绑定失败, status: 0x{command_status:08x}"

            return body, None
        except socket.timeout:
            return None, "接收响应超时"
        except Exception as e:
            return None, f"接收响应错误: {e}"

    def bind(self, system_id, password, bind_type='transceiver'):
        """绑定到 SMSC"""
        if not self.connect():
            return False, "连接失败"

        # 选择绑定类型
        bind_commands = {
            'transmitter': CMD_BIND_TRANSMITTER,
            'receiver': CMD_BIND_RECEIVER,
            'transceiver': CMD_BIND_TRANSCEIVER
        }

        if bind_type not in bind_commands:
            return False, f"无效的绑定类型: {bind_type}"

        command_id = bind_commands[bind_type]
        sequence_num = self._next_sequence()

        # 构建绑定请求体
        body = self._encode_cstring(system_id)
        body += self._encode_cstring(password)
        body += b'\x00'  # system_type (empty)
        body += struct.pack('>BBB', 0x34, 0x00, 0x00)  # interface_version, addr_ton, addr_npi
        body += b'\x00'  # address_range (empty)

        pdu = self._build_pdu(command_id, sequence_num, body)
        self.socket.send(pdu)

        # 读取响应
        expected_resp = command_id | 0x80000000
        _, error = self._read_response(expected_resp, sequence_num)

        if error:
            self.close()
            return False, error

        return True, f"绑定成功 ({bind_type})"

    def unbind(self):
        """解绑"""
        if not self.socket:
            return False, "未连接"

        sequence_num = self._next_sequence()
        pdu = self._build_pdu(CMD_UNBIND, sequence_num)
        self.socket.send(pdu)

        _, error = self._read_response(CMD_UNBIND_RESP, sequence_num)
        self.close()

        if error:
            return False, error

        return True, "解绑成功"

    def submit_sm(self, source_addr, dest_addr, message, encoding='GSM7'):
        """发送短消息"""
        if not self.socket:
            return False, "未连接"

        sequence_num = self._next_sequence()

        # 构建消息体
        body = b'\x00'  # service_type (empty)
        body += struct.pack('>BB', 0x01, 0x01)  # source_addr_ton, source_addr_npi
        body += self._encode_cstring(source_addr)
        body += struct.pack('>BB', 0x01, 0x01)  # dest_addr_ton, dest_addr_npi
        body += self._encode_cstring(dest_addr)
        body += struct.pack('>BBB', 0x00, 0x00, 0x00)  # esm_class, protocol_id, priority_flag
        body += b'\x00\x00'  # schedule_delivery_time, validity_period (empty)
        body += struct.pack('>BB', 0x00, 0x00)  # registered_delivery, replace_if_present_flag

        # 数据编码
        if encoding == 'UCS2':
            body += struct.pack('>BB', 0x08, 0x00)  # data_coding (UCS2), sm_default_msg_id
            encoded_msg = message.encode('utf-16-be')
        else:
            body += struct.pack('>BB', 0x00, 0x00)  # data_coding (GSM7), sm_default_msg_id
            encoded_msg = message.encode('ascii')

        body += struct.pack('>B', len(encoded_msg))  # sm_length
        body += encoded_msg

        pdu = self._build_pdu(CMD_SUBMIT_SM, sequence_num, body)
        self.socket.send(pdu)

        resp_body, error = self._read_response(CMD_SUBMIT_SM_RESP, sequence_num)

        if error:
            return False, error

        # 解析 message_id
        if resp_body:
            message_id = resp_body.split(b'\x00')[0].decode('ascii')
            return True, f"消息已发送, message_id: {message_id}"

        return True, "消息已发送"

    def test_full_flow(self, system_id, password, source_addr, dest_addr, message):
        """测试完整流程: 绑定 -> 发送消息 -> 解绑"""
        print(f"测试完整流程...")
        print(f"  服务器: {self.host}:{self.port}")
        print(f"  System ID: {system_id}")
        print(f"  源地址: {source_addr}")
        print(f"  目标地址: {dest_addr}")
        print(f"  消息: {message}")
        print()

        # 1. 绑定
        print("1. 绑定中...")
        success, msg = self.bind(system_id, password, 'transceiver')
        print(f"   {msg}")
        if not success:
            return False

        # 2. 发送消息
        print("2. 发送消息...")
        success, msg = self.submit_sm(source_addr, dest_addr, message)
        print(f"   {msg}")
        if not success:
            self.close()
            return False

        # 3. 解绑
        print("3. 解绑中...")
        success, msg = self.unbind()
        print(f"   {msg}")

        print()
        print("测试完成!")
        return True


def main():
    parser = argparse.ArgumentParser(description='简单 SMPP 客户端')
    parser.add_argument('--host', default='127.0.0.1', help='SMPP 服务器地址')
    parser.add_argument('--port', type=int, default=2775, help='SMPP 服务器端口')

    subparsers = parser.add_subparsers(dest='command', help='命令')

    # bind 命令
    bind_parser = subparsers.add_parser('bind', help='绑定到 SMSC')
    bind_parser.add_argument('--system-id', required=True, help='System ID')
    bind_parser.add_argument('--password', default='', help='密码')
    bind_parser.add_argument('--type', choices=['transmitter', 'receiver', 'transceiver'],
                            default='transceiver', help='绑定类型')

    # send 命令
    send_parser = subparsers.add_parser('send', help='发送消息')
    send_parser.add_argument('--system-id', required=True, help='System ID')
    send_parser.add_argument('--password', default='', help='密码')
    send_parser.add_argument('--from', dest='source', required=True, help='源地址')
    send_parser.add_argument('--to', dest='dest', required=True, help='目标地址')
    send_parser.add_argument('--message', required=True, help='消息内容')
    send_parser.add_argument('--encoding', choices=['GSM7', 'UCS2'], default='GSM7', help='编码')

    # unbind 命令
    unbind_parser = subparsers.add_parser('unbind', help='解绑')

    # test 命令 (完整流程测试)
    test_parser = subparsers.add_parser('test', help='运行完整流程测试')
    test_parser.add_argument('--system-id', default='test_user', help='System ID')
    test_parser.add_argument('--password', default='test_pass', help='密码')
    test_parser.add_argument('--from', dest='source', default='10086', help='源地址')
    test_parser.add_argument('--to', dest='dest', default='13800138000', help='目标地址')
    test_parser.add_argument('--message', default='Hello SMPP!', help='消息内容')

    args = parser.parse_args()

    if not args.command:
        parser.print_help()
        sys.exit(1)

    client = SMPPClient(args.host, args.port)

    if args.command == 'bind':
        success, msg = client.bind(args.system_id, args.password, args.type)
        print(msg)
        sys.exit(0 if success else 1)

    elif args.command == 'send':
        success, msg = client.bind(args.system_id, args.password, 'transceiver')
        if not success:
            print(f"绑定失败: {msg}")
            sys.exit(1)

        success, msg = client.submit_sm(args.source, args.dest, args.message, args.encoding)
        print(msg)

        client.unbind()
        sys.exit(0 if success else 1)

    elif args.command == 'unbind':
        success, msg = client.unbind()
        print(msg)
        sys.exit(0 if success else 1)

    elif args.command == 'test':
        client.test_full_flow(args.system_id, args.password, args.source, args.dest, args.message)
        sys.exit(0)


if __name__ == '__main__':
    main()
