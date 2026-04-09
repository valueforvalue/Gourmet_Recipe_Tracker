// File: lib/models/recipe.dart

class Recipe {
  final String title;
  final List<String> tags;
  final List<String> ingredients;
  final List<String> instructions;
  final String notes;

  const Recipe({
    required this.title,
    required this.tags,
    required this.ingredients,
    required this.instructions,
    required this.notes,
  });

  factory Recipe.fromJson(Map<String, dynamic> json) {
    return Recipe(
      title: json['title'] as String? ?? '',
      tags: List<String>.from(json['tags'] as List? ?? []),
      ingredients: List<String>.from(json['ingredients'] as List? ?? []),
      instructions: List<String>.from(json['instructions'] as List? ?? []),
      notes: json['notes'] as String? ?? '',
    );
  }

  Map<String, dynamic> toJson() => {
        'title': title,
        'tags': tags,
        'ingredients': ingredients,
        'instructions': instructions,
        'notes': notes,
      };

  Recipe copyWith({
    String? title,
    List<String>? tags,
    List<String>? ingredients,
    List<String>? instructions,
    String? notes,
  }) {
    return Recipe(
      title: title ?? this.title,
      tags: tags ?? this.tags,
      ingredients: ingredients ?? this.ingredients,
      instructions: instructions ?? this.instructions,
      notes: notes ?? this.notes,
    );
  }
}
