// File: lib/screens/recipe_detail_screen.dart
//
// Read-only view for a single recipe.  Provides an Edit button and PDF export.

import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';
import '../models/recipe.dart';
import 'recipe_edit_screen.dart';
import '../services/api_service.dart';

class RecipeDetailScreen extends StatelessWidget {
  final Recipe recipe;

  const RecipeDetailScreen({super.key, required this.recipe});

  Future<void> _openPdf(BuildContext context, {bool booklet = false}) async {
    try {
      final url = await ApiService.getRecipePdfUrl(recipe.title, booklet: booklet);
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
      appBar: AppBar(
        title: Text(recipe.title, overflow: TextOverflow.ellipsis),
        actions: [
          PopupMenuButton<String>(
            onSelected: (val) {
              if (val == 'pdf_letter') _openPdf(context);
              if (val == 'pdf_booklet') _openPdf(context, booklet: true);
            },
            itemBuilder: (_) => const [
              PopupMenuItem(value: 'pdf_letter', child: Text('Export PDF (Letter)')),
              PopupMenuItem(value: 'pdf_booklet', child: Text('Export PDF (Booklet)')),
            ],
          ),
          IconButton(
            icon: const Icon(Icons.edit),
            tooltip: 'Edit',
            onPressed: () {
              Navigator.pushReplacement(
                context,
                MaterialPageRoute(builder: (_) => RecipeEditScreen(recipe: recipe)),
              );
            },
          ),
        ],
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          // Tags
          if (recipe.tags.isNotEmpty) ...[
            Wrap(
              spacing: 6,
              children: recipe.tags.map((t) => Chip(label: Text(t))).toList(),
            ),
            const SizedBox(height: 16),
          ],

          // Ingredients
          _SectionHeader(title: 'Ingredients'),
          const SizedBox(height: 8),
          ...recipe.ingredients.map(
            (ing) => Padding(
              padding: const EdgeInsets.symmetric(vertical: 2),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text('• ', style: TextStyle(fontSize: 16)),
                  Expanded(child: Text(ing, style: const TextStyle(fontSize: 16))),
                ],
              ),
            ),
          ),
          const SizedBox(height: 20),

          // Instructions
          _SectionHeader(title: 'Instructions'),
          const SizedBox(height: 8),
          ...recipe.instructions.asMap().entries.map(
            (e) => Padding(
              padding: const EdgeInsets.symmetric(vertical: 4),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text('${e.key + 1}. ', style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
                  Expanded(child: Text(e.value, style: const TextStyle(fontSize: 16))),
                ],
              ),
            ),
          ),

          // Notes
          if (recipe.notes.isNotEmpty) ...[
            const SizedBox(height: 20),
            _SectionHeader(title: 'Notes'),
            const SizedBox(height: 8),
            Text(recipe.notes, style: const TextStyle(fontStyle: FontStyle.italic, fontSize: 15)),
          ],
        ],
      ),
    );
  }
}

class _SectionHeader extends StatelessWidget {
  final String title;
  const _SectionHeader({required this.title});

  @override
  Widget build(BuildContext context) {
    return Text(
      title,
      style: Theme.of(context).textTheme.titleLarge?.copyWith(fontWeight: FontWeight.bold),
    );
  }
}
