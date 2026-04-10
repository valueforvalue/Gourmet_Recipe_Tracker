// File: lib/screens/settings_screen.dart
//
// Lets the user set the server base URL and test connectivity.

import 'package:flutter/material.dart';
import '../services/api_service.dart';

class SettingsScreen extends StatefulWidget {
  const SettingsScreen({super.key});

  @override
  State<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends State<SettingsScreen> {
  final _urlCtrl = TextEditingController();
  bool _testing = false;
  String? _statusMessage;
  bool? _statusOk;

  @override
  void initState() {
    super.initState();
    _loadUrl();
  }

  @override
  void dispose() {
    _urlCtrl.dispose();
    super.dispose();
  }

  Future<void> _loadUrl() async {
    final url = await ApiService.getBaseUrl();
    setState(() => _urlCtrl.text = url ?? '');
  }

  Future<void> _save() async {
    await ApiService.setBaseUrl(_urlCtrl.text.trim());
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Server URL saved')),
      );
    }
  }

  Future<void> _test() async {
    await ApiService.setBaseUrl(_urlCtrl.text.trim());
    setState(() {
      _testing = true;
      _statusMessage = null;
      _statusOk = null;
    });
    final result = await ApiService.testConnection();
    setState(() {
      _testing = false;
      _statusOk = result.ok;
      _statusMessage = result.ok
          ? 'Connected successfully!'
          : 'Could not reach server: ${result.error}';
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Server Settings')),
      body: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Server URL', style: Theme.of(context).textTheme.titleMedium),
            const SizedBox(height: 4),
            const Text(
              'Enter the address of your home server, e.g.\n'
              'http://192.168.1.100:8080',
              style: TextStyle(color: Colors.grey),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: _urlCtrl,
              decoration: const InputDecoration(
                border: OutlineInputBorder(),
                hintText: 'http://192.168.1.100:8080',
              ),
              keyboardType: TextInputType.url,
            ),
            const SizedBox(height: 16),
            Row(
              children: [
                ElevatedButton(onPressed: _save, child: const Text('Save')),
                const SizedBox(width: 12),
                _testing
                    ? const SizedBox(width: 36, height: 36, child: CircularProgressIndicator(strokeWidth: 2))
                    : OutlinedButton.icon(
                        onPressed: _test,
                        icon: const Icon(Icons.wifi_find),
                        label: const Text('Test Connection'),
                      ),
              ],
            ),
            if (_statusMessage != null) ...[
              const SizedBox(height: 16),
              Row(
                children: [
                  Icon(
                    _statusOk! ? Icons.check_circle : Icons.error,
                    color: _statusOk! ? Colors.green : Colors.red,
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      _statusMessage!,
                      style: TextStyle(color: _statusOk! ? Colors.green : Colors.red),
                    ),
                  ),
                ],
              ),
            ],
          ],
        ),
      ),
    );
  }
}
