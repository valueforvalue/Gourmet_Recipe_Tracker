// File: lib/screens/settings_screen.dart
//
// Lets the user set the server base URL, test connectivity, and view version info.

import 'package:flutter/material.dart';
import 'package:package_info_plus/package_info_plus.dart';
import '../services/api_service.dart';
import '../theme/season.dart';

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
  String _appVersion = '';
  String _serverVersion = '';

  @override
  void initState() {
    super.initState();
    _loadUrl();
    _loadVersions();
  }

  @override
  void dispose() {
    _urlCtrl.dispose();
    super.dispose();
  }

  Future<void> _loadUrl() async {
    final url = await ApiService.getBaseUrl();
    if (mounted) setState(() => _urlCtrl.text = url ?? '');
  }

  Future<void> _loadVersions() async {
    try {
      final info = await PackageInfo.fromPlatform();
      if (mounted) setState(() => _appVersion = '${info.version} (build ${info.buildNumber})');
    } catch (_) {
      if (mounted) setState(() => _appVersion = 'unknown');
    }
    try {
      final v = await ApiService.getServerVersion();
      if (mounted) setState(() => _serverVersion = '${v.appVersion} (schema ${v.schemaVersion})');
    } catch (_) {
      if (mounted) setState(() => _serverVersion = 'not connected');
    }
  }

  Future<void> _save() async {
    await ApiService.setBaseUrl(_urlCtrl.text.trim());
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Server URL saved')),
      );
    }
    _loadVersions();
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
    _loadVersions();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Server Settings')),
      body: SingleChildScrollView(
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
            const SizedBox(height: 24),
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text('About', style: Theme.of(context).textTheme.titleMedium),
                    const SizedBox(height: 8),
                    _VersionRow(label: 'App version', value: _appVersion.isEmpty ? '…' : _appVersion),
                    const SizedBox(height: 4),
                    _VersionRow(label: 'Server', value: _serverVersion.isEmpty ? '…' : _serverVersion),
                    const SizedBox(height: 4),
                    _VersionRow(label: 'Season', value: activeSeason.value.label),
                    const SizedBox(height: 16),
                    Text('Theme preview', style: Theme.of(context).textTheme.titleSmall),
                    const SizedBox(height: 8),
                    DropdownButton<Season>(
                      value: activeSeason.value,
                      isExpanded: true,
                      items: Season.values
                          .map((s) => DropdownMenuItem(value: s, child: Text(s.label)))
                          .toList(),
                      onChanged: (s) {
                        if (s != null) setState(() => activeSeason.value = s);
                      },
                    ),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _VersionRow extends StatelessWidget {
  final String label;
  final String value;
  const _VersionRow({required this.label, required this.value});

  @override
  Widget build(BuildContext context) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        SizedBox(width: 110, child: Text(label, style: const TextStyle(color: Colors.grey))),
        Expanded(child: Text(value)),
      ],
    );
  }
}
