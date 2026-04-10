// File: lib/main.dart
//
// App entry point.  Uses a bottom navigation bar to switch between
// the recipe list and the master cookbook export screen.

import 'package:flutter/material.dart';
import 'screens/recipe_list_screen.dart';
import 'screens/cookbook_screen.dart';

void main() {
  WidgetsFlutterBinding.ensureInitialized();
  runApp(const RecipeTrackerApp());
}

class RecipeTrackerApp extends StatelessWidget {
  const RecipeTrackerApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Recipe Tracker',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.brown),
        useMaterial3: true,
      ),
      home: const _MainShell(),
    );
  }
}

class _MainShell extends StatefulWidget {
  const _MainShell();

  @override
  State<_MainShell> createState() => _MainShellState();
}

class _MainShellState extends State<_MainShell> {
  int _selectedIndex = 0;

  static const _screens = [
    RecipeListScreen(),
    CookbookScreen(),
  ];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: _screens[_selectedIndex],
      bottomNavigationBar: NavigationBar(
        selectedIndex: _selectedIndex,
        onDestinationSelected: (i) => setState(() => _selectedIndex = i),
        destinations: const [
          NavigationDestination(icon: Icon(Icons.restaurant_menu), label: 'Recipes'),
          NavigationDestination(icon: Icon(Icons.menu_book), label: 'Cookbook'),
        ],
      ),
    );
  }
}
