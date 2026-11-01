// File: lib/screens/deleted_screen.dart
//
// Lists soft-deleted recipes with restore + permanent-purge actions.

import 'package:flutter/material.dart';
import '../models/recipe.dart';
import '../services/api_service.dart';

class DeletedScreen extends StatefulWidget {
  const DeletedScreen({super.key});

  @override
  State<DeletedScreen> createState() => _DeletedScreenState();
}

class _DeletedScreenState extends State<DeletedScreen> {
  List<Recipe> _deleted = [];
  bool _loading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final recipes = await ApiService.getDeletedRecipes();
      setState(() {
        _deleted = recipes;
        _loading = false;
      });
    } catch (e) {
      setState(() {
        _error = e.toString();
        _loading = false;
      });
    }
  }

  Future<void> _restore(Recipe recipe) async {
    try {
      await ApiService.restoreRecipe(recipe.title);
      _load();
    } catch (e) {
      _snack('Error: $e');
    }
  }

  Future<void> _purge(Recipe recipe) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (_) => AlertDialog(
        title: const Text('Delete Forever?'),
        content: Text('"${recipe.title}" will be permanently removed. This cannot be undone.'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context, false), child: const Text('Cancel')),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Delete Forever', style: TextStyle(color: Colors.red)),
          ),
        ],
      ),
    );
    if (confirmed == true) {
      try {
        await ApiService.purgeRecipe(recipe.title);
        _load();
      } catch (e) {
        _snack('Error: $e');
      }
    }
  }

  Future<void> _purgeAll() async {
    final controller = TextEditingController();
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (_) => AlertDialog(
        title: const Text('Purge All Deleted?'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('This permanently deletes every deleted recipe. Type "yesdeary" to confirm.'),
            const SizedBox(height: 12),
            TextField(
              controller: controller,
              decoration: const InputDecoration(border: OutlineInputBorder(), hintText: 'yesdeary'),
            ),
          ],
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context, false), child: const Text('Cancel')),
          TextButton(
            onPressed: () => Navigator.pop(context, controller.text.trim() == 'yesdeary'),
            child: const Text('Purge All', style: TextStyle(color: Colors.red)),
          ),
        ],
      ),
    );
    if (confirmed == true) {
      try {
        await ApiService.purgeAllDeleted();
        _load();
      } catch (e) {
        _snack('Error: $e');
      }
    }
  }

  void _snack(String msg) {
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(msg)));
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Deleted Recipes'),
        actions: [
          if (_deleted.isNotEmpty)
            IconButton(
              icon: const Icon(Icons.delete_forever, color: Colors.red),
              tooltip: 'Purge All',
              onPressed: _purgeAll,
            ),
        ],
      ),
      body: _buildBody(),
    );
  }

  Widget _buildBody() {
    if (_loading) return const Center(child: CircularProgressIndicator());
    if (_error != null) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(Icons.wifi_off, size: 64, color: Colors.grey),
            const SizedBox(height: 16),
            Text(_error!, textAlign: TextAlign.center, style: const TextStyle(color: Colors.grey)),
            const SizedBox(height: 16),
            TextButton(onPressed: _load, child: const Text('Retry')),
          ],
        ),
      );
    }
    if (_deleted.isEmpty) {
      return const Center(child: Text('No deleted recipes.', style: TextStyle(color: Colors.grey)));
    }
    return ListView.builder(
      itemCount: _deleted.length,
      itemBuilder: (_, i) {
        final recipe = _deleted[i];
        return ListTile(
          title: Text(recipe.title),
          subtitle: recipe.tags.isEmpty ? null : Text(recipe.tags.join(', ')),
          trailing: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              IconButton(icon: const Icon(Icons.restore), tooltip: 'Restore', onPressed: () => _restore(recipe)),
              IconButton(
                icon: const Icon(Icons.delete_forever, color: Colors.red),
                tooltip: 'Delete Forever',
                onPressed: () => _purge(recipe),
              ),
            ],
          ),
        );
      },
    );
  }
}
