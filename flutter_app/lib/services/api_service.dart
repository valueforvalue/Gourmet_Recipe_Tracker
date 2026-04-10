// File: lib/services/api_service.dart
//
// All network calls go through this class.  The base URL is loaded from
// SharedPreferences so the user can point the app at their home server.

import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';
import '../models/recipe.dart';

class ApiService {
  static const String _baseUrlKey = 'server_base_url';
  static const String _defaultBaseUrl = 'http://10.0.2.2:8080';

  // ── Settings ──────────────────────────────────────────────────────────────

  static Future<String> getBaseUrl() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString(_baseUrlKey) ?? _defaultBaseUrl;
  }

  static Future<void> setBaseUrl(String url) async {
    final prefs = await SharedPreferences.getInstance();
    // Strip trailing slash for consistency
    await prefs.setString(_baseUrlKey, url.endsWith('/') ? url.substring(0, url.length - 1) : url);
  }

  // ── Recipe CRUD ───────────────────────────────────────────────────────────

  /// Returns every non-deleted recipe, sorted alphabetically.
  static Future<List<Recipe>> getAllRecipes() async {
    final base = await getBaseUrl();
    final response = await http
        .get(Uri.parse('$base/api/recipes'))
        .timeout(const Duration(seconds: 15));
    if (response.statusCode == 200) {
      final List<dynamic> data = json.decode(response.body) as List<dynamic>;
      return data.map((e) => Recipe.fromJson(e as Map<String, dynamic>)).toList();
    }
    throw Exception('Failed to load recipes (${response.statusCode})');
  }

  /// Creates or updates a recipe (server performs an UPSERT on title).
  static Future<void> saveRecipe(Recipe recipe) async {
    final base = await getBaseUrl();
    final response = await http
        .post(
          Uri.parse('$base/api/save'),
          headers: {'Content-Type': 'application/json'},
          body: json.encode(recipe.toJson()),
        )
        .timeout(const Duration(seconds: 15));
    if (response.statusCode != 200) {
      throw Exception('Failed to save recipe (${response.statusCode})');
    }
  }

  /// Soft-deletes a recipe by title.
  static Future<void> deleteRecipe(String title) async {
    final base = await getBaseUrl();
    final response = await http
        .post(
          Uri.parse('$base/api/delete?title=${Uri.encodeComponent(title)}'),
        )
        .timeout(const Duration(seconds: 15));
    if (response.statusCode != 200) {
      throw Exception('Failed to delete recipe (${response.statusCode})');
    }
  }

  // ── Web Scraper ───────────────────────────────────────────────────────────

  /// Asks the server to scrape a recipe URL and returns a pre-filled Recipe.
  static Future<Recipe> scrapeUrl(String url) async {
    final base = await getBaseUrl();
    final response = await http
        .get(Uri.parse('$base/api/scrape?url=${Uri.encodeComponent(url)}'))
        .timeout(const Duration(seconds: 30));
    if (response.statusCode == 200) {
      return Recipe.fromJson(json.decode(response.body) as Map<String, dynamic>);
    }
    throw Exception('Failed to scrape URL (${response.statusCode}): ${response.body}');
  }

  // ── PDF Export ────────────────────────────────────────────────────────────

  /// Returns the URL to download a single-recipe PDF.
  static Future<String> getRecipePdfUrl(String title, {bool booklet = false}) async {
    final base = await getBaseUrl();
    return '$base/api/export/pdf?title=${Uri.encodeComponent(title)}&booklet=$booklet';
  }

  /// Returns the URL to download the master cookbook PDF.
  static Future<String> getCookbookPdfUrl({bool booklet = false}) async {
    final base = await getBaseUrl();
    return '$base/api/export/cookbook?booklet=$booklet';
  }

  // ── Health check ──────────────────────────────────────────────────────────

  /// Returns true if the server responds within 5 seconds; on failure returns
  /// false and rethrows the underlying error so callers can surface a message.
  static Future<({bool ok, String error})> testConnection() async {
    try {
      final base = await getBaseUrl();
      final response = await http
          .get(Uri.parse('$base/api/recipes'))
          .timeout(const Duration(seconds: 5));
      return (ok: response.statusCode == 200, error: response.statusCode == 200 ? '' : 'Server returned HTTP ${response.statusCode}');
    } catch (e) {
      return (ok: false, error: e.toString());
    }
  }
}
