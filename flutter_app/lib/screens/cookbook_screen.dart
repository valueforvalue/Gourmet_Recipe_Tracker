// File: lib/screens/cookbook_screen.dart
//
// Lets the user trigger the master cookbook PDF export on the server.

import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';
import '../services/api_service.dart';

class CookbookScreen extends StatelessWidget {
  const CookbookScreen({super.key});

  Future<void> _export(BuildContext context, {bool booklet = false}) async {
    try {
      final url = await ApiService.getCookbookPdfUrl(booklet: booklet);
      if (!await launchUrl(Uri.parse(url), mode: LaunchMode.externalApplication)) {
        if (context.mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('Could not open PDF')),
          );
        }
      }
    } catch (e) {
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Error: $e')));
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Master Cookbook')),
      body: Center(
        child: Padding(
          padding: const EdgeInsets.all(32),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Icon(Icons.menu_book, size: 80, color: Colors.brown),
              const SizedBox(height: 24),
              Text(
                'Export Your Cookbook',
                style: Theme.of(context).textTheme.headlineSmall,
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 8),
              const Text(
                'Compile every recipe into a single PDF, grouped and alphabetised by category.',
                textAlign: TextAlign.center,
                style: TextStyle(color: Colors.grey),
              ),
              const SizedBox(height: 32),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton.icon(
                  icon: const Icon(Icons.picture_as_pdf),
                  label: const Text('Export (Letter 8.5" × 11")'),
                  onPressed: () => _export(context),
                ),
              ),
              const SizedBox(height: 12),
              SizedBox(
                width: double.infinity,
                child: OutlinedButton.icon(
                  icon: const Icon(Icons.picture_as_pdf),
                  label: const Text('Export (Booklet 5.5" × 8.5")'),
                  onPressed: () => _export(context, booklet: true),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
