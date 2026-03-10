module Main exposing (..)

import Browser
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick, onInput)
import Http
import Json.Decode as Decode
import Json.Encode as Encode



-- 1. MODEL


type View
    = EntryView
    | ListView


type alias Recipe =
    { title : String, tags : List String, notes : String }


type alias Model =
    { currentView : View
    , recipes : List Recipe
    , filterText : String
    , title : String
    , tags : String
    , ingredients : String
    , instructions : String
    , notes : String
    , status : String
    }


initialModel : Model
initialModel =
    { currentView = EntryView
    , recipes = []
    , filterText = ""
    , title = ""
    , tags = ""
    , ingredients = ""
    , instructions = ""
    , notes = ""
    , status = "Ready"
    }



-- 2. UPDATE


type Msg
    = SetView View
    | UpdateFilter String
    | UpdateTitle String
    | UpdateTags String
    | UpdateIngredients String
    | UpdateInstructions String
    | UpdateNotes String
    | SaveRecipe
    | RecipeSaved (Result Http.Error ())
    | FetchRecipes
    | RecipesFetched (Result Http.Error (List Recipe))
    | DeleteRecipe String
    | Deleted (Result Http.Error ())


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        SetView viewType ->
            ( { model | currentView = viewType, status = "" }
            , if viewType == ListView then
                fetchRecipes

              else
                Cmd.none
            )

        UpdateFilter val ->
            ( { model | filterText = val }, Cmd.none )

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
                    ( { initialModel | status = "Saved!" }, Cmd.none )

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

        DeleteRecipe title ->
            ( { model | status = "Deleting..." }, deleteRequest title )

        Deleted _ ->
            ( model, fetchRecipes )



-- 3. VIEW


view : Model -> Html Msg
view model =
    div [ style "background-color" "#FDF5E6", style "min-height" "100vh", style "font-family" "serif" ]
        [ div [ style "background" "#6B705C", style "padding" "10px", style "display" "flex", style "justify-content" "center", style "gap" "15px", style "position" "sticky", style "top" "0", style "z-index" "10" ]
            [ button (onClick (SetView EntryView) :: navStyle) [ text "Add" ]
            , button (onClick (SetView ListView) :: navStyle) [ text "Recipe Box" ]
            ]
        , div [ style "padding" "20px", style "max-width" "600px", style "margin" "0 auto" ]
            [ case model.currentView of
                EntryView ->
                    viewEntryForm model

                ListView ->
                    viewRecipeList model
            , p [ style "text-align" "center", style "color" "#6B705C" ] [ text model.status ]
            ]
        ]


navStyle =
    [ style "background" "none", style "border" "none", style "color" "white", style "font-size" "18px", style "cursor" "pointer" ]


viewEntryForm : Model -> Html Msg
viewEntryForm model =
    div [ style "background" "white", style "padding" "20px", style "border-radius" "8px", style "box-shadow" "0 2px 10px rgba(0,0,0,0.1)" ]
        [ h2 [ style "color" "#6B705C" ] [ text "New Recipe" ]
        , input (placeholder "Title" :: value model.title :: onInput UpdateTitle :: inputStyle) []
        , input (placeholder "Tags" :: value model.tags :: onInput UpdateTags :: inputStyle) []
        , textarea (placeholder "Ingredients" :: rows 6 :: value model.ingredients :: onInput UpdateIngredients :: inputStyle) []
        , textarea (placeholder "Instructions" :: rows 6 :: value model.instructions :: onInput UpdateInstructions :: inputStyle) []
        , input (placeholder "Notes" :: value model.notes :: onInput UpdateNotes :: inputStyle) []
        , button [ onClick SaveRecipe, style "background" "#A5A58D", style "color" "white", style "padding" "15px", style "width" "100%", style "border" "none", style "border-radius" "4px" ] [ text "Save Recipe" ]
        ]


viewRecipeList : Model -> Html Msg
viewRecipeList model =
    let
        filteredRecipes =
            List.filter (\r -> String.contains (String.toLower model.filterText) (String.toLower r.title)) model.recipes
    in
    div []
        [ h2 [ style "color" "#6B705C" ] [ text "Recipe Box" ]
        , input (placeholder "Search your recipes..." :: value model.filterText :: onInput UpdateFilter :: inputStyle) []
        , div [ style "margin-bottom" "20px", style "display" "flex", style "gap" "10px" ]
            [ a (href "/api/export/cookbook?booklet=false" :: target "_blank" :: masterBtnStyle "#6B705C") [ text "Letter Cookbook" ]
            , a (href "/api/export/cookbook?booklet=true" :: target "_blank" :: masterBtnStyle "#A5A58D") [ text "Booklet Cookbook" ]
            ]
        , div [] (List.map viewRecipeCard filteredRecipes)
        ]


viewRecipeCard : Recipe -> Html Msg
viewRecipeCard recipe =
    div [ style "background" "white", style "padding" "15px", style "margin-bottom" "15px", style "border-left" "6px solid #6B705C", style "border-radius" "4px", style "box-shadow" "0 1px 3px rgba(0,0,0,0.1)" ]
        [ strong [ style "font-size" "18px", style "display" "block" ] [ text recipe.title ]
        , p [ style "font-size" "13px", style "color" "#777", style "margin" "5px 0" ] [ text (String.join ",  " recipe.tags) ]
        , div [ style "display" "flex", style "gap" "8px", style "margin-top" "10px", style "flex-wrap" "wrap" ]
            [ a (href ("/api/export/pdf?title=" ++ recipe.title ++ "&booklet=false") :: target "_blank" :: printBtnStyle "#6B705C") [ text "Letter" ]
            , a (href ("/api/export/pdf?title=" ++ recipe.title ++ "&booklet=true") :: target "_blank" :: printBtnStyle "#A5A58D") [ text "Booklet" ]
            , button [ onClick (DeleteRecipe recipe.title), style "background" "#e5e5e5", style "color" "#a00", style "border" "none", style "padding" "8px", style "border-radius" "4px", style "font-size" "12px", style "cursor" "pointer" ] [ text "Delete" ]
            ]
        ]


masterBtnStyle c =
    [ style "background" c, style "color" "white", style "text-decoration" "none", style "padding" "10px", style "flex" "1", style "text-align" "center", style "border-radius" "4px", style "font-size" "11px" ]


printBtnStyle c =
    [ style "background" c, style "color" "white", style "text-decoration" "none", style "padding" "8px", style "flex" "1", style "text-align" "center", style "border-radius" "4px", style "font-size" "12px" ]


inputStyle =
    [ style "width" "100%", style "margin-bottom" "10px", style "padding" "12px", style "border" "1px solid #DDD", style "border-radius" "4px", style "box-sizing" "border-box" ]



-- HTTP


fetchRecipes =
    Http.get { url = "/api/recipes", expect = Http.expectJson RecipesFetched (Decode.map3 Recipe (Decode.field "title" Decode.string) (Decode.field "tags" (Decode.list Decode.string)) (Decode.field "notes" Decode.string) |> Decode.list) }


postRecipe model =
    Http.post { url = "/api/save", body = Http.jsonBody (Encode.object [ ( "title", Encode.string model.title ), ( "tags", Encode.list Encode.string (String.split "," model.tags |> List.map String.trim |> List.filter (not << String.isEmpty)) ), ( "ingredients", Encode.list Encode.string (String.split "\n" model.ingredients) ), ( "instructions", Encode.list Encode.string (String.split "\n" model.instructions) ), ( "notes", Encode.string model.notes ) ]), expect = Http.expectWhatever RecipeSaved }


deleteRequest title =
    Http.post { url = "/api/delete?title=" ++ title, body = Http.emptyBody, expect = Http.expectWhatever Deleted }


main : Program () Model Msg
main =
    Browser.element { init = \_ -> ( initialModel, Cmd.none ), view = view, update = update, subscriptions = \_ -> Sub.none }
