module Main exposing (..)

import Browser
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick, onInput)
import Http
import Json.Decode as Decode
import Json.Encode as Encode
import Regex



-- 1. MODEL


type View
    = EntryView
    | ListView
    | ReadingView Recipe


type alias Recipe =
    { title : String
    , tags : List String
    , ingredients : List String
    , instructions : List String
    , notes : String
    }


type alias Model =
    { currentView : View
    , recipes : List Recipe
    , filterText : String
    , scrapeUrl : String
    , title : String
    , tags : String
    , ingredients : String
    , instructions : String
    , notes : String
    , status : String
    , deletingTitle : Maybe String
    }


initialModel : Model
initialModel =
    { currentView = EntryView
    , recipes = []
    , filterText = ""
    , scrapeUrl = ""
    , title = ""
    , tags = ""
    , ingredients = ""
    , instructions = ""
    , notes = ""
    , status = "Ready"
    , deletingTitle = Nothing
    }



-- 2. UPDATE


type Msg
    = SetView View
    | UpdateFilter String
    | UpdateScrapeUrl String
    | RunScrape
    | ScrapeResult (Result Http.Error Recipe)
    | ClearForm
    | UpdateTitle String
    | UpdateTags String
    | UpdateIngredients String
    | UpdateInstructions String
    | UpdateNotes String
    | SaveRecipe
    | RecipeSaved (Result Http.Error ())
    | FetchRecipes
    | RecipesFetched (Result Http.Error (List Recipe))
    | ConfirmDelete String
    | CancelDelete
    | ExecuteDelete String
    | Deleted (Result Http.Error ())
    | EditRecipe Recipe
    | CancelEdit
    | OpenReader Recipe


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        SetView viewType ->
            ( { model | currentView = viewType, status = "", deletingTitle = Nothing }
            , if viewType == ListView then
                fetchRecipes

              else
                Cmd.none
            )

        UpdateFilter val ->
            ( { model | filterText = val }, Cmd.none )

        UpdateScrapeUrl val ->
            ( { model | scrapeUrl = val }, Cmd.none )

        RunScrape ->
            ( { model | status = "Scraping recipe data..." }, scrapeRecipe model.scrapeUrl )

        ScrapeResult res ->
            case res of
                Ok r ->
                    ( { model
                        | title = r.title
                        , ingredients = String.join "\n" r.ingredients
                        , instructions = String.join "\n" r.instructions
                        , notes = r.notes
                        , status = "Imported! Please review."
                      }
                    , Cmd.none
                    )

                Err _ ->
                    ( { model | status = "Failed to scrape URL." }, Cmd.none )

        ClearForm ->
            ( { model
                | title = ""
                , tags = ""
                , ingredients = ""
                , instructions = ""
                , notes = ""
                , scrapeUrl = ""
                , status = "Form Cleared."
              }
            , Cmd.none
            )

        UpdateTitle val ->
            ( { model | title = val }, Cmd.none )

        UpdateTags val ->
            ( { model | tags = val }, Cmd.none )

        UpdateIngredients val ->
            ( { model | ingredients = val }, Cmd.none )

        UpdateInstructions val ->
            ( { model | instructions = val }, Cmd.none )

        UpdateNotes val ->
            ( { model | notes = val }, Cmd.none )

        SaveRecipe ->
            ( { model | status = "Saving to records..." }, postRecipe model )

        RecipeSaved res ->
            case res of
                Ok _ ->
                    ( { initialModel | status = "Saved Successfully!", currentView = ListView }, fetchRecipes )

                Err _ ->
                    ( { model | status = "Error saving." }, Cmd.none )

        FetchRecipes ->
            ( model, fetchRecipes )

        RecipesFetched res ->
            case res of
                Ok list ->
                    ( { model | recipes = list }, Cmd.none )

                Err _ ->
                    ( model, Cmd.none )

        ConfirmDelete title ->
            ( { model | deletingTitle = Just title }, Cmd.none )

        CancelDelete ->
            ( { model | deletingTitle = Nothing }, Cmd.none )

        ExecuteDelete title ->
            ( { model | status = "Deleting...", deletingTitle = Nothing }, deleteRequest title )

        Deleted _ ->
            ( model, fetchRecipes )

        EditRecipe recipe ->
            ( { model
                | currentView = EntryView
                , title = recipe.title
                , tags = String.join ", " recipe.tags
                , ingredients = String.join "\n" recipe.ingredients
                , instructions = String.join "\n" recipe.instructions
                , notes = recipe.notes
                , status = "Editing: " ++ recipe.title
              }
            , Cmd.none
            )

        CancelEdit ->
            ( initialModel, Cmd.none )

        OpenReader recipe ->
            ( { model | currentView = ReadingView recipe, status = "" }, Cmd.none )



-- 3. VIEW


view : Model -> Html Msg
view model =
    div [ style "background-color" "#FDF5E6", style "min-height" "100vh", style "font-family" "serif" ]
        [ -- NAVIGATION BAR
          div [ style "background" "#6B705C", style "padding" "10px", style "display" "flex", style "align-items" "center", style "position" "sticky", style "top" "0", style "z-index" "10", style "box-shadow" "0 2px 5px rgba(0,0,0,0.2)" ]
            [ div [ style "flex" "1" ] []
            , div [ style "display" "flex", style "gap" "20px" ]
                [ button (onClick (SetView EntryView) :: navStyle) [ text "Add" ]
                , button (onClick (SetView ListView) :: navStyle) [ text "Recipe Box" ]
                ]
            , div [ style "flex" "1", style "text-align" "right", style "padding-right" "15px" ]
                [ span [ style "color" "#fff", style "font-style" "italic", style "font-size" "14px", style "opacity" "0.8" ] [ text "Morris Family" ] ]
            ]
        , div [ style "padding" "20px", style "max-width" "600px", style "margin" "0 auto" ]
            [ case model.currentView of
                EntryView ->
                    viewEntryForm model

                ListView ->
                    viewRecipeList model

                ReadingView recipe ->
                    viewReader recipe
            , p [ style "text-align" "center", style "color" "#6B705C", style "font-weight" "bold" ] [ text model.status ]
            ]
        ]


viewEntryForm : Model -> Html Msg
viewEntryForm model =
    div []
        [ -- SCRAPER SECTION
          div [ style "background" "#f0f0e0", style "padding" "15px", style "border-radius" "8px", style "margin-bottom" "20px", style "border" "1px dashed #6B705C" ]
            [ h4 [ style "margin-top" "0", style "color" "#6B705C" ] [ text "✨ Import from Web" ]
            , div [ style "display" "flex", style "gap" "10px" ]
                [ input [ placeholder "Paste recipe URL here...", value model.scrapeUrl, onInput UpdateScrapeUrl, style "flex" "3", style "padding" "10px", style "border-radius" "4px", style "border" "1px solid #ccc" ] []
                , button [ onClick RunScrape, style "flex" "1", style "background" "#6B705C", style "color" "white", style "border" "none", style "border-radius" "4px", style "cursor" "pointer" ] [ text "Import" ]
                ]
            ]
        , -- MANUAL ENTRY SECTION
          div [ style "background" "white", style "padding" "20px", style "border-radius" "8px", style "box-shadow" "0 2px 10px rgba(0,0,0,0.1)" ]
            [ div [ style "display" "flex", style "justify-content" "space-between", style "align-items" "center" ]
                [ h2 [ style "color" "#6B705C" ]
                    [ text
                        (if String.startsWith "Editing" model.status then
                            "Edit Recipe"

                         else
                            "Morris Recipe Entry"
                        )
                    ]
                , button [ onClick ClearForm, style "background" "none", style "border" "none", style "color" "#a00", style "text-decoration" "underline", style "cursor" "pointer", style "font-size" "12px" ] [ text "Clear All" ]
                ]
            , input (placeholder "Title" :: value model.title :: onInput UpdateTitle :: inputStyle) []
            , input (placeholder "Tags (e.g. Dinner, Beef, Slow Cooker)" :: value model.tags :: onInput UpdateTags :: inputStyle) []
            , textarea (placeholder "Ingredients (One per line)" :: rows 6 :: value model.ingredients :: onInput UpdateIngredients :: inputStyle) []
            , textarea (placeholder "Instructions (One per line)" :: rows 6 :: value model.instructions :: onInput UpdateInstructions :: inputStyle) []
            , input (placeholder "Notes" :: value model.notes :: onInput UpdateNotes :: inputStyle) []
            , div [ style "display" "flex", style "gap" "10px" ]
                [ button [ onClick SaveRecipe, style "background" "#A5A58D", style "color" "white", style "padding" "15px", style "flex" "2", style "border" "none", style "border-radius" "4px", style "cursor" "pointer" ] [ text "Save Recipe" ]
                , if String.startsWith "Editing" model.status then
                    button [ onClick CancelEdit, style "background" "#e5e5e5", style "color" "#555", style "padding" "15px", style "flex" "1", style "border" "none", style "border-radius" "4px", style "cursor" "pointer" ] [ text "Cancel" ]

                  else
                    text ""
                ]
            ]
        ]


viewRecipeList : Model -> Html Msg
viewRecipeList model =
    let
        searchLower =
            String.toLower model.filterText

        filteredRecipes =
            List.filter
                (\r ->
                    String.contains searchLower (String.toLower r.title)
                        || List.any (\ing -> String.contains searchLower (String.toLower ing)) r.ingredients
                )
                model.recipes

        verse =
            getDailyVerse model
    in
    div []
        [ h2 [ style "color" "#6B705C", style "margin-bottom" "5px" ] [ text "Morris Family Recipe Box" ]
        , -- SCRIPTURE CARD
          div [ style "margin-bottom" "25px", style "padding" "20px", style "background" "#fdfaf0", style "border-radius" "12px", style "border" "1px solid #e9e2d0", style "box-shadow" "0 2px 4px rgba(0,0,0,0.05)" ]
            [ p [ style "font-style" "italic", style "margin" "0", style "color" "#4a4a4a", style "line-height" "1.5", style "font-size" "16px" ] [ text ("\"" ++ verse.text ++ "\"") ]
            , p [ style "font-size" "13px", style "text-align" "right", style "margin" "10px 0 0 0", style "font-weight" "bold", style "color" "#6B705C" ] [ text verse.ref ]
            ]
        , input (placeholder "Search title or ingredient..." :: value model.filterText :: onInput UpdateFilter :: inputStyle) []
        , -- MASTER COOKBOOK BUTTONS
          div [ style "margin-bottom" "20px", style "display" "flex", style "gap" "10px" ]
            [ a (href "/api/export/cookbook?booklet=false" :: target "_blank" :: masterBtnStyle "#6B705C") [ text "Letter Cookbook" ]
            , a (href "/api/export/cookbook?booklet=true" :: target "_blank" :: masterBtnStyle "#A5A58D") [ text "Booklet Cookbook" ]
            ]
        , div [] (List.map (viewRecipeCard model.deletingTitle) filteredRecipes)
        ]


viewRecipeCard : Maybe String -> Recipe -> Html Msg
viewRecipeCard deletingTitle recipe =
    let
        isDeleting =
            deletingTitle == Just recipe.title
    in
    div [ style "background" "white", style "padding" "15px", style "margin-bottom" "15px", style "border-left" "6px solid #6B705C", style "border-radius" "4px", style "box-shadow" "0 1px 3px rgba(0,0,0,0.1)" ]
        [ strong [ style "font-size" "20px", style "display" "block", style "cursor" "pointer", style "color" "#6B705C", style "text-decoration" "underline", onClick (OpenReader recipe) ] [ text recipe.title ]
        , p [ style "font-size" "13px", style "color" "#777", style "margin" "5px 0" ] [ text (String.join ",  " recipe.tags) ]
        , if isDeleting then
            div [ style "background" "#fff0f0", style "padding" "10px", style "border-radius" "4px", style "margin-top" "10px", style "display" "flex", style "align-items" "center", style "justify-content" "space-between" ]
                [ span [ style "color" "#a00" ] [ text "Delete?" ]
                , div [ style "display" "flex", style "gap" "10px" ]
                    [ button [ onClick (ExecuteDelete recipe.title), style "background" "#a00", style "color" "white", style "border" "none", style "padding" "5px 15px", style "border-radius" "4px", style "cursor" "pointer" ] [ text "Yes" ]
                    , button [ onClick CancelDelete, style "background" "#ccc", style "border" "none", style "padding" "5px 15px", style "border-radius" "4px", style "cursor" "pointer" ] [ text "No" ]
                    ]
                ]

          else
            div [ style "display" "flex", style "gap" "8px", style "margin-top" "10px", style "flex-wrap" "wrap" ]
                [ a (href ("/api/export/pdf?title=" ++ recipe.title ++ "&booklet=false") :: target "_blank" :: actionBtnStyle "#6B705C") [ text "Letter" ]
                , a (href ("/api/export/pdf?title=" ++ recipe.title ++ "&booklet=true") :: target "_blank" :: actionBtnStyle "#A5A58D") [ text "Booklet" ]
                , button (onClick (EditRecipe recipe) :: actionBtnStyle "#d4a373") [ text "Edit" ]
                , button (onClick (ConfirmDelete recipe.title) :: actionBtnStyle "#e5e5e5" ++ [ style "color" "#a00" ]) [ text "Delete" ]
                ]
        ]


viewReader : Recipe -> Html Msg
viewReader recipe =
    div [ style "background" "white", style "padding" "30px", style "border-radius" "8px", style "box-shadow" "0 4px 20px rgba(0,0,0,0.1)" ]
        [ h1 [ style "color" "#6B705C", style "margin-bottom" "10px" ] [ text recipe.title ]
        , div [ style "font-style" "italic", style "color" "#888", style "margin-bottom" "20px" ] [ text (String.join ",  " recipe.tags) ]
        , h3 [ style "border-bottom" "2px solid #A5A58D", style "padding-bottom" "5px" ] [ text "Ingredients" ]
        , ul [ style "line-height" "1.8", style "font-size" "18px" ] (List.map (\ing -> li [] [ text ing ]) recipe.ingredients)
        , h3 [ style "border-bottom" "2px solid #A5A58D", style "padding-bottom" "5px", style "margin-top" "30px" ] [ text "Instructions" ]
        , ol [ style "line-height" "1.6", style "font-size" "18px" ] (List.map (\inst -> li [ style "margin-bottom" "15px" ] [ text (cleanInstruction inst) ]) recipe.instructions)
        , if String.isEmpty recipe.notes then
            text ""

          else
            div [ style "margin-top" "30px", style "padding" "15px", style "background" "#f9f9f9", style "border-radius" "4px" ]
                [ h4 [ style "margin" "0 0 10px 0" ] [ text "Notes" ], p [] [ text recipe.notes ] ]
        , button [ onClick (SetView ListView), style "margin-top" "30px", style "width" "100%", style "padding" "15px", style "background" "#6B705C", style "color" "white", style "border" "none", style "border-radius" "4px", style "cursor" "pointer" ] [ text "Done Reading" ]
        ]



-- 4. UTILS & SCRIPTURE


bibleVerses : List { text : String, ref : String }
bibleVerses =
    [ { text = "Go, eat your bread with joy, and drink your wine with a merry heart, for God has already approved what you do.", ref = "Ecclesiastes 9:7" }
    , { text = "Whether you eat or drink, or whatever you do, do all to the glory of God.", ref = "1 Corinthians 10:31" }
    , { text = "He fills your soul with good things.", ref = "Psalm 103:5" }
    , { text = "Better is a dinner of herbs where love is than a fattened ox and hatred with it.", ref = "Proverbs 15:17" }
    , { text = "Give us this day our daily bread.", ref = "Matthew 6:11" }
    , { text = "Taste and see that the Lord is good.", ref = "Psalm 34:8" }
    ]


getDailyVerse : Model -> { text : String, ref : String }
getDailyVerse model =
    let
        seed =
            String.length model.filterText + List.length model.recipes

        index =
            modBy (List.length bibleVerses) seed
    in
    bibleVerses |> List.drop index |> List.head |> Maybe.withDefault { text = "", ref = "" }


cleanInstruction : String -> String
cleanInstruction inst =
    let
        re =
            Regex.fromString "^\\d+[\\.\\)]\\s*" |> Maybe.withDefault Regex.never
    in
    Regex.replace re (\_ -> "") (String.trim inst)



-- 5. HTTP & JSON


scrapeRecipe : String -> Cmd Msg
scrapeRecipe url =
    Http.get { url = "/api/scrape?url=" ++ url, expect = Http.expectJson ScrapeResult recipeDecoder }


fetchRecipes : Cmd Msg
fetchRecipes =
    Http.get { url = "/api/recipes", expect = Http.expectJson RecipesFetched (Decode.list recipeDecoder) }


recipeDecoder : Decode.Decoder Recipe
recipeDecoder =
    Decode.map5 Recipe (Decode.field "title" Decode.string) (Decode.field "tags" (Decode.list Decode.string)) (Decode.field "ingredients" (Decode.list Decode.string)) (Decode.field "instructions" (Decode.list Decode.string)) (Decode.field "notes" Decode.string)


postRecipe : Model -> Cmd Msg
postRecipe model =
    Http.post
        { url = "/api/save"
        , body =
            Http.jsonBody
                (Encode.object
                    [ ( "title", Encode.string model.title )
                    , ( "ingredients", Encode.list Encode.string (String.split "\n" model.ingredients |> List.map String.trim |> List.filter (not << String.isEmpty)) )
                    , ( "instructions", Encode.list Encode.string (String.split "\n" model.instructions |> List.map String.trim |> List.filter (not << String.isEmpty)) )
                    , ( "tags", Encode.list Encode.string (String.split "," model.tags |> List.map String.trim |> List.filter (not << String.isEmpty)) )
                    , ( "notes", Encode.string model.notes )
                    ]
                )
        , expect = Http.expectWhatever RecipeSaved
        }


deleteRequest : String -> Cmd Msg
deleteRequest title =
    Http.post { url = "/api/delete?title=" ++ title, body = Http.emptyBody, expect = Http.expectWhatever Deleted }



-- STYLES


inputStyle =
    [ style "width" "100%", style "margin-bottom" "12px", style "padding" "10px", style "border" "1px solid #ccc", style "border-radius" "4px", style "box-sizing" "border-box" ]


actionBtnStyle c =
    [ style "background" c, style "border" "none", style "padding" "8px 15px", style "border-radius" "4px", style "cursor" "pointer", style "color" "white", style "text-decoration" "none", style "text-align" "center", style "font-size" "12px" ]


masterBtnStyle c =
    [ style "background" c, style "color" "white", style "text-decoration" "none", style "padding" "10px", style "flex" "1", style "text-align" "center", style "border-radius" "4px", style "font-size" "11px" ]


navStyle =
    [ style "background" "none", style "border" "none", style "color" "white", style "font-size" "18px", style "cursor" "pointer", style "font-weight" "bold" ]



-- 6. MAIN


main : Program () Model Msg
main =
    Browser.element
        { init = \() -> ( initialModel, Cmd.none )
        , view = view
        , update = update
        , subscriptions = \_ -> Sub.none
        }
