// File: lib/screens/recipe_detail_screen.dart
//
// Read-only view for a single recipe: scaling, PDF export, and Edit.

import 'package:android_intent_plus/android_intent.dart';
import 'package:android_intent_plus/flag.dart';
import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';
import '../models/recipe.dart';
import 'recipe_edit_screen.dart';
import '../services/api_service.dart';

class RecipeDetailScreen extends StatefulWidget {
  final Recipe recipe;
  const RecipeDetailScreen({super.key, required this.recipe});

  @override
  State<RecipeDetailScreen> createState() => _RecipeDetailScreenState();
}

class _RecipeDetailScreenState extends State<RecipeDetailScreen> {
  Recipe? _scaled;
  String _factor = '1';

  Recipe get _display => _scaled ?? widget.recipe;
  bool get _isScaled => _scaled != null;

  static const _scaleOptions = [
    (label: '½×', factor: '0.5'),
    (label: '1×', factor: '1'),
    (label: '2×', factor: '2'),
    (label: '3×', factor: '3'),
  ];

  Future<void> _scale(String factor) async {
    setState(() => _factor = factor);
    if (factor == '1') {
      setState(() => _scaled = null);
      return;
    }
    try {
      final scaled = await ApiService.scaleRecipe(widget.recipe.title, factor);
      if (mounted) setState(() => _scaled = scaled);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Failed to scale: $e')));
      }
    }
  }

  void _reset() {
    setState(() {
      _scaled = null;
      _factor = '1';
    });
  }

  Future<void> _openPdf(BuildContext context, {bool booklet = false}) async {
    try {
      final url = await ApiService.getRecipePdfUrl(widget.recipe.title, booklet: booklet);
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

  Future<void> _setTimer() async {
    try {
      const intent = AndroidIntent(
        action: 'android.intent.action.SET_TIMER',
        flags: [Flag.FLAG_ACTIVITY_NEW_TASK],
      );
      await intent.launch();
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Could not open timer: $e')));
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    return Scaffold(
      appBar: AppBar(
        title: Text(widget.recipe.title, overflow: TextOverflow.ellipsis),
        actions: [
          IconButton(
            icon: const Icon(Icons.timer_outlined),
            tooltip: 'Set Timer',
            onPressed: _setTimer,
          ),
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
                MaterialPageRoute(builder: (_) => RecipeEditScreen(recipe: widget.recipe)),
              );
            },
          ),
        ],
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          // Tags
          if (widget.recipe.tags.isNotEmpty) ...[
            Wrap(
              spacing: 6,
              children: widget.recipe.tags.map((t) => Chip(label: Text(t))).toList(),
            ),
            const SizedBox(height: 16),
          ],

          // Scale controls
          Wrap(
            spacing: 8,
            children: [
              for (final o in _scaleOptions)
                ChoiceChip(
                  label: Text(o.label),
                  selected: _factor == o.factor,
                  onSelected: (_) => _scale(o.factor),
                ),
            ],
          ),
          if (_isScaled) ...[
            const SizedBox(height: 12),
            Card(
              color: scheme.tertiaryContainer,
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
                child: Row(
                  children: [
                    Expanded(
                      child: Text(
                        '⚠ ×$_factor — temporarily scaled',
                        style: TextStyle(fontWeight: FontWeight.bold, color: scheme.onTertiaryContainer),
                      ),
                    ),
                    TextButton(onPressed: _reset, child: const Text('Reset')),
                  ],
                ),
              ),
            ),
          ],
          const SizedBox(height: 16),

          // Ingredients
          _SectionHeader(title: 'Ingredients'),
          const SizedBox(height: 8),
          ..._display.ingredients.map(
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
          ..._display.instructions.asMap().entries.map(
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
          if (_display.notes.isNotEmpty) ...[
            const SizedBox(height: 20),
            _SectionHeader(title: 'Notes'),
            const SizedBox(height: 8),
            _LinkifiedText(text: _display.notes),
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

// Renders text with URLs as tappable links (line breaks preserved).
class _LinkifiedText extends StatefulWidget {
  final String text;
  const _LinkifiedText({required this.text});

  @override
  State<_LinkifiedText> createState() => _LinkifiedTextState();
}

class _LinkifiedTextState extends State<_LinkifiedText> {
  static final RegExp _urlRe = RegExp(r'https?://[^\s]+');
  final List<TapGestureRecognizer> _recognizers = [];

  @override
  void dispose() {
    for (final r in _recognizers) {
      r.dispose();
    }
    super.dispose();
  }

  Future<void> _open(String url) async {
    await launchUrl(Uri.parse(url), mode: LaunchMode.externalApplication);
  }

  @override
  Widget build(BuildContext context) {
    final baseStyle = const TextStyle(fontStyle: FontStyle.italic, fontSize: 15);
    final linkColor = Theme.of(context).colorScheme.primary;
    final spans = <TextSpan>[];
    final matches = _urlRe.allMatches(widget.text).toList();

    var cursor = 0;
    for (final m in matches) {
      if (m.start > cursor) {
        spans.add(TextSpan(text: widget.text.substring(cursor, m.start)));
      }
      final recognizer = TapGestureRecognizer()..onTap = () => _open(m.group(0)!);
      _recognizers.add(recognizer);
      spans.add(TextSpan(
        text: m.group(0),
        style: baseStyle.copyWith(color: linkColor, decoration: TextDecoration.underline),
        recognizer: recognizer,
      ));
      cursor = m.end;
    }
    if (cursor < widget.text.length) {
      spans.add(TextSpan(text: widget.text.substring(cursor)));
    }

    return Text.rich(TextSpan(style: baseStyle, children: spans));
  }
}
