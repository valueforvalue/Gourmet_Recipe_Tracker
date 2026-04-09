// File: test/widget_test.dart
//
// Basic smoke test: verifies the app launches and shows the bottom nav bar.

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:morris_recipe_tracker/main.dart';

void main() {
  testWidgets('App launches and shows navigation bar', (WidgetTester tester) async {
    await tester.pumpWidget(const RecipeTrackerApp());

    // Bottom nav destinations should be visible immediately.
    expect(find.text('Recipes'), findsOneWidget);
    expect(find.text('Cookbook'), findsOneWidget);
  });
}
