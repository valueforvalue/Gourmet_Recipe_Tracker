// File: lib/screens/recipe_list_screen.dart
//
// Home screen — displays all recipes with live search/filter.

import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';
import '../models/recipe.dart';
import '../services/api_service.dart';
import '../services/daily_verse.dart';
import 'recipe_detail_screen.dart';
import 'recipe_edit_screen.dart';
import 'settings_screen.dart';

class RecipeListScreen extends StatefulWidget {
  const RecipeListScreen({super.key});

  @override
  State<RecipeListScreen> createState() => _RecipeListScreenState();
}

class _RecipeListScreenState extends State<RecipeListScreen> {
  List<Recipe> _allRecipes = [];
  List<Recipe> _filtered = [];
  bool _loading = true;
  String? _error;
  final TextEditingController _searchCtrl = TextEditingController();
  String _activeTag = '';

  @override
  void initState() {
    super.initState();
    _loadRecipes();
    _searchCtrl.addListener(_applyFilter);
  }

  @override
  void dispose() {
    _searchCtrl.dispose();
    super.dispose();
  }

  Future<void> _loadRecipes() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    // Don't attempt a network call if no server has been configured yet.
    final url = await ApiService.getBaseUrl();
    if (url == null) {
      setState(() {
        _loading = false;
        _error = 'No server URL configured. Open Settings to set one.';
      });
      return;
    }
    try {
      final recipes = await ApiService.getAllRecipes();
      setState(() {
        _allRecipes = recipes;
        _loading = false;
      });
      _applyFilter();
    } catch (e) {
      setState(() {
        _error = e.toString();
        _loading = false;
      });
    }
  }

  void _applyFilter() {
    final query = _searchCtrl.text.toLowerCase();
    setState(() {
      _filtered = _allRecipes.where((r) {
        final matchesTag = _activeTag.isEmpty || r.tags.contains(_activeTag);
        final matchesQuery = query.isEmpty ||
            r.title.toLowerCase().contains(query) ||
            r.tags.any((t) => t.toLowerCase().contains(query)) ||
            r.ingredients.any((i) => i.toLowerCase().contains(query));
        return matchesTag && matchesQuery;
      }).toList();
    });
  }

  Set<String> get _allTags {
    final tags = <String>{};
    for (final r in _allRecipes) {
      tags.addAll(r.tags);
    }
    return tags;
  }

  int _tagCount(String tag) => _allRecipes.where((r) => r.tags.contains(tag)).length;

  String get _totalLabel {
    final n = _allRecipes.length;
    return n == 1 ? '1 recipe' : '$n recipes';
  }

  Future<void> _deleteRecipe(Recipe recipe) async {
    final confirm = await showDialog<bool>(
      context: context,
      builder: (_) => AlertDialog(
        title: const Text('Delete Recipe?'),
        content: Text('Are you sure you want to delete "${recipe.title}"?'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context, false), child: const Text('Cancel')),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Delete', style: TextStyle(color: Colors.red)),
          ),
        ],
      ),
    );
    if (confirm == true) {
      try {
        await ApiService.deleteRecipe(recipe.title);
        _loadRecipes();
      } catch (e) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Error: $e')));
        }
      }
    }
  }

  Future<void> _openFeedback() async {
    const url = 'https://submit-form.com/SftusnkmF';
    if (!await launchUrl(Uri.parse(url), mode: LaunchMode.externalApplication)) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Could not open feedback form')),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: const [
            Text('Recipe Tracker'),
            Text('Morris Family', style: TextStyle(fontSize: 12, fontStyle: FontStyle.italic, color: Colors.white70)),
          ],
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.feedback_outlined),
            tooltip: 'Send Feedback',
            onPressed: _openFeedback,
          ),
          IconButton(
            icon: const Icon(Icons.settings),
            tooltip: 'Server Settings',
            onPressed: () async {
              await Navigator.push(context, MaterialPageRoute(builder: (_) => const SettingsScreen()));
              _loadRecipes();
            },
          ),
          IconButton(
            icon: const Icon(Icons.refresh),
            tooltip: 'Refresh',
            onPressed: _loadRecipes,
          ),
        ],
        bottom: PreferredSize(
          preferredSize: const Size.fromHeight(56),
          child: Padding(
            padding: const EdgeInsets.fromLTRB(12, 0, 12, 8),
            child: TextField(
              controller: _searchCtrl,
              decoration: InputDecoration(
                hintText: 'Search recipes…',
                prefixIcon: const Icon(Icons.search),
                filled: true,
                fillColor: Theme.of(context).colorScheme.surface,
                border: OutlineInputBorder(borderRadius: BorderRadius.circular(24), borderSide: BorderSide.none),
                contentPadding: EdgeInsets.zero,
              ),
            ),
          ),
        ),
      ),
      body: _buildBody(),
      floatingActionButton: FloatingActionButton(
        tooltip: 'Add Recipe',
        onPressed: () async {
          await Navigator.push(
            context,
            MaterialPageRoute(builder: (_) => const RecipeEditScreen()),
          );
          _loadRecipes();
        },
        child: const Icon(Icons.add),
      ),
    );
  }

  Widget _buildBody() {
    if (_loading) return const Center(child: CircularProgressIndicator());
    if (_error != null) {
      final isUnconfigured = _error!.contains('No server URL');
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(isUnconfigured ? Icons.settings_ethernet : Icons.wifi_off,
                  size: 64, color: Colors.grey),
              const SizedBox(height: 16),
              Text(
                isUnconfigured ? 'No server configured' : 'Could not reach server',
                style: Theme.of(context).textTheme.titleLarge,
              ),
              const SizedBox(height: 8),
              Text(_error!, textAlign: TextAlign.center, style: const TextStyle(color: Colors.grey)),
              const SizedBox(height: 24),
              ElevatedButton.icon(
                onPressed: () async {
                  await Navigator.push(context, MaterialPageRoute(builder: (_) => const SettingsScreen()));
                  _loadRecipes();
                },
                icon: const Icon(Icons.settings),
                label: Text(isUnconfigured ? 'Set Server URL' : 'Update Server URL'),
              ),
              if (!isUnconfigured) ...[
                const SizedBox(height: 8),
                TextButton(onPressed: _loadRecipes, child: const Text('Retry')),
              ],
            ],
          ),
        ),
      );
    }

    final tags = _allTags.toList()..sort();
    final verse = verseForToday();

    return Column(
      children: [
        Card(
          margin: const EdgeInsets.fromLTRB(12, 0, 12, 8),
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text('"${verse.text}"', style: const TextStyle(fontStyle: FontStyle.italic)),
                const SizedBox(height: 8),
                Align(
                  alignment: Alignment.centerRight,
                  child: Text(verse.ref, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 13, color: Colors.grey)),
                ),
              ],
            ),
          ),
        ),
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 0, 16, 4),
          child: Align(
            alignment: Alignment.centerLeft,
            child: Text(_totalLabel, style: const TextStyle(color: Colors.grey, fontSize: 13)),
          ),
        ),
        if (tags.isNotEmpty)
          SizedBox(
            height: 44,
            child: ListView(
              scrollDirection: Axis.horizontal,
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
              children: [
                _TagChip(label: 'All', selected: _activeTag.isEmpty, onTap: () {
                  setState(() => _activeTag = '');
                  _applyFilter();
                }),
                for (final tag in tags)
                  _TagChip(label: '$tag (${_tagCount(tag)})', selected: _activeTag == tag, onTap: () {
                    setState(() => _activeTag = _activeTag == tag ? '' : tag);
                    _applyFilter();
                  }),
              ],
            ),
          ),
        Expanded(
          child: _filtered.isEmpty
              ? const Center(child: Text('No recipes found.', style: TextStyle(color: Colors.grey)))
              : ListView.builder(
                  padding: const EdgeInsets.only(bottom: 88),
                  itemCount: _filtered.length,
                  itemBuilder: (_, i) {
                    final recipe = _filtered[i];
                    return ListTile(
                      title: Text(recipe.title),
                      subtitle: recipe.tags.isEmpty ? null : Text(recipe.tags.join(', ')),
                      trailing: IconButton(
                        icon: const Icon(Icons.delete_outline, color: Colors.red),
                        tooltip: 'Delete',
                        onPressed: () => _deleteRecipe(recipe),
                      ),
                      onTap: () async {
                        await Navigator.push(
                          context,
                          MaterialPageRoute(builder: (_) => RecipeDetailScreen(recipe: recipe)),
                        );
                        _loadRecipes();
                      },
                    );
                  },
                ),
        ),
      ],
    );
  }
}

class _TagChip extends StatelessWidget {
  final String label;
  final bool selected;
  final VoidCallback onTap;

  const _TagChip({required this.label, required this.selected, required this.onTap});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(right: 6),
      child: FilterChip(
        label: Text(label),
        selected: selected,
        onSelected: (_) => onTap(),
      ),
    );
  }
}
