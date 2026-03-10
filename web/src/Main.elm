module Main exposing (..)

import Browser
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick, onInput)
import Http
import Json.Encode as Encode



-- 1. MODEL


type alias Model =
    { title : String
    , tags : String
    , ingredients : String
    , instructions : String
    , notes : String
    , status : String
    }


initialModel : Model
initialModel =
    { title = ""
    , tags = ""
    , ingredients = ""
    , instructions = ""
    , notes = ""
    , status = "Ready"
    }



-- 2. UPDATE


type Msg
    = UpdateTitle String
    | UpdateTags String
    | UpdateIngredients String
    | UpdateInstructions String
    | UpdateNotes String
    | SaveRecipe
    | RecipeSaved (Result Http.Error ())


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
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
            ( { model | status = "Saving..." }, postRecipe model )

        RecipeSaved res ->
            case res of
                Ok _ ->
                    ( { initialModel | status = "Successfully Saved!" }, Cmd.none )

                Err _ ->
                    ( { model | status = "Error: Could not save." }, Cmd.none )



-- 3. VIEW


view : Model -> Html Msg
view model =
    div [ style "padding" "40px", style "font-family" "sans-serif", style "max-width" "600px" ]
        [ h1 [] [ text "Gourmet Recipe Entry" ]
        , div [ style "margin-bottom" "10px" ] [ input [ placeholder "Title", value model.title, onInput UpdateTitle, style "width" "100%" ] [] ]
        , div [ style "margin-bottom" "10px" ] [ input [ placeholder "Tags (Comma separated)", value model.tags, onInput UpdateTags, style "width" "100%" ] [] ]
        , div [ style "margin-bottom" "10px" ] [ textarea [ placeholder "Ingredients (Line by line)", rows 10, value model.ingredients, onInput UpdateIngredients, style "width" "100%" ] [] ]
        , div [ style "margin-bottom" "10px" ] [ textarea [ placeholder "Instructions (Line by line)", rows 10, value model.instructions, onInput UpdateInstructions, style "width" "100%" ] [] ]
        , div [ style "margin-bottom" "10px" ] [ input [ placeholder "Notes", value model.notes, onInput UpdateNotes, style "width" "100%" ] [] ]
        , button [ onClick SaveRecipe, style "background" "#4CAF50", style "color" "white", style "padding" "10px 20px", style "border" "none", style "cursor" "pointer" ] [ text "Save Recipe" ]
        , p [] [ text model.status ]
        ]



-- 4. HTTP / JSON


postRecipe : Model -> Cmd Msg
postRecipe model =
    Http.post
        { url = "/api/save"
        , body = Http.jsonBody (encodeRecipe model)
        , expect = Http.expectWhatever RecipeSaved
        }


encodeRecipe : Model -> Encode.Value
encodeRecipe model =
    Encode.object
        [ ( "title", Encode.string model.title )
        , ( "tags", Encode.list Encode.string (String.split "," model.tags |> List.map String.trim |> List.filter (not << String.isEmpty)) )
        , ( "ingredients", Encode.list Encode.string (String.split "\n" model.ingredients |> List.filter (not << String.isEmpty)) )
        , ( "instructions", Encode.list Encode.string (String.split "\n" model.instructions |> List.filter (not << String.isEmpty)) )
        , ( "notes", Encode.string model.notes )
        ]


main : Program () Model Msg
main =
    Browser.element
        { init = \_ -> ( initialModel, Cmd.none )
        , view = view
        , update = update
        , subscriptions = \_ -> Sub.none
        }
