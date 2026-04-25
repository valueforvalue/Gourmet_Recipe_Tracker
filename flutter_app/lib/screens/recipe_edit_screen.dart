// File: lib/screens/recipe_edit_screen.dart
//
// Create or edit a recipe.  Supports manual entry and server-side URL scraping.

import 'package:flutter/material.dart';
import '../models/recipe.dart';
import '../services/api_service.dart';

class RecipeEditScreen extends StatefulWidget {
  /// Pass an existing recipe to edit it; omit to create a new one.
  final Recipe? recipe;

  const RecipeEditScreen({super.key, this.recipe});

  @override
  State<RecipeEditScreen> createState() => _RecipeEditScreenState();
}

class _RecipeEditScreenState extends State<RecipeEditScreen> {
  final _formKey = GlobalKey<FormState>();
  final _titleCtrl = TextEditingController();
  final _notesCtrl = TextEditingController();
  final _scrapeUrlCtrl = TextEditingController();

  List<TextEditingController> _ingredientCtrls = [];
  List<TextEditingController> _instructionCtrls = [];
  List<TextEditingController> _tagCtrls = [];

  bool _saving = false;
  bool _scraping = false;

  @override
  void initState() {
    super.initState();
    final r = widget.recipe;
    if (r != null) {
      _titleCtrl.text = r.title;
      _notesCtrl.text = r.notes;
      _ingredientCtrls = r.ingredients.map((s) => TextEditingController(text: s)).toList();
      _instructionCtrls = r.instructions.map((s) => TextEditingController(text: s)).toList();
      _tagCtrls = r.tags.map((s) => TextEditingController(text: s)).toList();
    }
    if (_ingredientCtrls.isEmpty) _ingredientCtrls.add(TextEditingController());
    if (_instructionCtrls.isEmpty) _instructionCtrls.add(TextEditingController());
    if (_tagCtrls.isEmpty) _tagCtrls.add(TextEditingController());
  }

  @override
  void dispose() {
    _titleCtrl.dispose();
    _notesCtrl.dispose();
    _scrapeUrlCtrl.dispose();
    for (final c in [..._ingredientCtrls, ..._instructionCtrls, ..._tagCtrls]) {
      c.dispose();
    }
    super.dispose();
  }

  // ── Scraping ──────────────────────────────────────────────────────────────

  Future<void> _scrapeUrl() async {
    final url = _scrapeUrlCtrl.text.trim();
    if (url.isEmpty) return;
    setState(() => _scraping = true);
    try {
      final scraped = await ApiService.scrapeUrl(url);
      setState(() {
        _titleCtrl.text = scraped.title;
        _notesCtrl.text = scraped.notes;
        _ingredientCtrls = scraped.ingredients.isEmpty
            ? [TextEditingController()]
            : scraped.ingredients.map((s) => TextEditingController(text: s)).toList();
        _instructionCtrls = scraped.instructions.isEmpty
            ? [TextEditingController()]
            : scraped.instructions.map((s) => TextEditingController(text: s)).toList();
        _tagCtrls = scraped.tags.isEmpty
            ? [TextEditingController()]
            : scraped.tags.map((s) => TextEditingController(text: s)).toList();
      });
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Recipe scraped — review and save.')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Scrape failed: $e')));
      }
    } finally {
      setState(() => _scraping = false);
    }
  }

  // ── Save ──────────────────────────────────────────────────────────────────

  Future<void> _save() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() => _saving = true);
    try {
      final recipe = Recipe(
        title: _titleCtrl.text.trim(),
        notes: _notesCtrl.text.trim(),
        ingredients: _ingredientCtrls.map((c) => c.text.trim()).where((s) => s.isNotEmpty).toList(),
        instructions: _instructionCtrls.map((c) => c.text.trim()).where((s) => s.isNotEmpty).toList(),
        tags: _tagCtrls.map((c) => c.text.trim()).where((s) => s.isNotEmpty).toList(),
      );
      await ApiService.saveRecipe(recipe);
      if (mounted) Navigator.pop(context);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Save failed: $e')));
      }
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  // ── Helpers ───────────────────────────────────────────────────────────────

  void _addController(List<TextEditingController> list) {
    setState(() => list.add(TextEditingController()));
  }

  void _removeController(List<TextEditingController> list, int index) {
    if (list.length <= 1) return; // keep at least one row
    setState(() {
      list[index].dispose();
      list.removeAt(index);
    });
  }

  // ── Build ─────────────────────────────────────────────────────────────────

  @override
  Widget build(BuildContext context) {
    final isEditing = widget.recipe != null;
    return Scaffold(
      appBar: AppBar(
        title: Text(isEditing ? 'Edit Recipe' : 'New Recipe'),
        actions: [
          if (_saving)
            const Padding(
              padding: EdgeInsets.all(16),
              child: SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white)),
            )
          else
            IconButton(icon: const Icon(Icons.save), tooltip: 'Save', onPressed: _save),
        ],
      ),
      body: Form(
        key: _formKey,
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            // ── Import from URL ──────────────────────────────────────────
            Card(
              child: Padding(
                padding: const EdgeInsets.all(12),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text('Import from URL', style: Theme.of(context).textTheme.titleSmall),
                    const SizedBox(height: 8),
                    Row(
                      children: [
                        Expanded(
                          child: TextFormField(
                            controller: _scrapeUrlCtrl,
                            decoration: const InputDecoration(
                              hintText: 'https://…',
                              border: OutlineInputBorder(),
                              isDense: true,
                            ),
                            keyboardType: TextInputType.url,
                          ),
                        ),
                        const SizedBox(width: 8),
                        _scraping
                            ? const SizedBox(width: 36, height: 36, child: CircularProgressIndicator(strokeWidth: 2))
                            : ElevatedButton(onPressed: _scrapeUrl, child: const Text('Scrape')),
                      ],
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),

            // ── Title ────────────────────────────────────────────────────
            TextFormField(
              controller: _titleCtrl,
              decoration: const InputDecoration(labelText: 'Title *', border: OutlineInputBorder()),
              validator: (v) => (v == null || v.trim().isEmpty) ? 'Title is required' : null,
            ),
            const SizedBox(height: 16),

            // ── Tags ─────────────────────────────────────────────────────
            _SectionLabel(label: 'Tags'),
            ..._tagCtrls.asMap().entries.map((e) => _ListRow(
                  ctrl: e.value,
                  hint: 'e.g. Dinner',
                  onRemove: () => _removeController(_tagCtrls, e.key),
                )),
            _AddButton(label: 'Add Tag', onPressed: () => _addController(_tagCtrls)),
            const SizedBox(height: 16),

            // ── Ingredients ──────────────────────────────────────────────
            _SectionLabel(label: 'Ingredients'),
            ..._ingredientCtrls.asMap().entries.map((e) => _ListRow(
                  ctrl: e.value,
                  hint: 'e.g. 2 cups flour',
                  onRemove: () => _removeController(_ingredientCtrls, e.key),
                )),
            _AddButton(label: 'Add Ingredient', onPressed: () => _addController(_ingredientCtrls)),
            const SizedBox(height: 16),

            // ── Instructions ─────────────────────────────────────────────
            _SectionLabel(label: 'Instructions'),
            ..._instructionCtrls.asMap().entries.map((e) => _ListRow(
                  ctrl: e.value,
                  hint: 'Step ${e.key + 1}',
                  onRemove: () => _removeController(_instructionCtrls, e.key),
                  maxLines: 3,
                )),
            _AddButton(label: 'Add Step', onPressed: () => _addController(_instructionCtrls)),
            const SizedBox(height: 16),

            // ── Notes ────────────────────────────────────────────────────
            TextFormField(
              controller: _notesCtrl,
              decoration: const InputDecoration(labelText: 'Notes', border: OutlineInputBorder()),
              maxLines: 4,
            ),
            const SizedBox(height: 80),
          ],
        ),
      ),
    );
  }
}

// ── Small helper widgets ───────────────────────────────────────────────────

class _SectionLabel extends StatelessWidget {
  final String label;
  const _SectionLabel({required this.label});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Text(label, style: Theme.of(context).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold)),
    );
  }
}

class _ListRow extends StatelessWidget {
  final TextEditingController ctrl;
  final String hint;
  final VoidCallback onRemove;
  final int maxLines;

  const _ListRow({
    required this.ctrl,
    required this.hint,
    required this.onRemove,
    this.maxLines = 1,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Row(
        children: [
          Expanded(
            child: TextFormField(
              controller: ctrl,
              decoration: InputDecoration(hintText: hint, border: const OutlineInputBorder(), isDense: true),
              maxLines: maxLines,
            ),
          ),
          IconButton(icon: const Icon(Icons.remove_circle_outline, color: Colors.red), onPressed: onRemove),
        ],
      ),
    );
  }
}

class _AddButton extends StatelessWidget {
  final String label;
  final VoidCallback onPressed;

  const _AddButton({required this.label, required this.onPressed});

  @override
  Widget build(BuildContext context) {
    return TextButton.icon(
      onPressed: onPressed,
      icon: const Icon(Icons.add),
      label: Text(label),
    );
  }
}
