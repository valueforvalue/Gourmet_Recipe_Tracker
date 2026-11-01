module Main exposing (..)

import Browser
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onCheck, onClick, onInput)
import Http
import Json.Decode as Decode
import Json.Encode as Encode
import Regex
import Set
import Task
import Time
import Url



-- 1. MODEL


type View
    = EntryView
    | ListView
    | ReadingView Recipe
    | DeletedView
    | FeedbackView


type PurgeIntent
    = PurgeSelected
    | PurgeAll


type Season
    = Spring
    | Summer
    | Autumn
    | Winter


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
    , activeTag : String
    , scrapeUrl : String
    , title : String
    , tags : String
    , ingredients : String
    , instructions : String
    , notes : String
    , status : String
    , deletingTitle : Maybe String
    , deleted : List Recipe
    , selectedDeleted : List String
    , purgeIntent : Maybe PurgeIntent
    , purgeConfirm : String
    , verseIndex : Int
    , season : Season
    , scaled : Maybe Recipe
    , scaleFactor : String
    }


initialModel : Model
initialModel =
    { currentView = EntryView
    , recipes = []
    , filterText = ""
    , activeTag = ""
    , scrapeUrl = ""
    , title = ""
    , tags = ""
    , ingredients = ""
    , instructions = ""
    , notes = ""
    , status = "Ready"
    , deletingTitle = Nothing
    , deleted = []
    , selectedDeleted = []
    , purgeIntent = Nothing
    , purgeConfirm = ""
    , verseIndex = 0
    , season = Spring
    , scaled = Nothing
    , scaleFactor = "1"
    }



-- 2. UPDATE


type Msg
    = SetView View
    | UpdateFilter String
    | SetActiveTag String
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
    | FetchDeleted
    | DeletedFetched (Result Http.Error (List Recipe))
    | ToggleDeletedSelect String
    | ToggleSelectAll
    | RestoreSelected
    | Restored (Result Http.Error ())
    | StartPurge PurgeIntent
    | UpdatePurgeConfirm String
    | CancelPurge
    | ConfirmPurge
    | Purged (Result Http.Error ())
    | ScaleRecipe String
    | Scaled (Result Http.Error Recipe)
    | ResetScale
    | GotTime Time.Posix


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        SetView viewType ->
            ( { model | currentView = viewType, status = "", deletingTitle = Nothing, purgeIntent = Nothing, scaled = Nothing, scaleFactor = "1" }
            , if viewType == ListView then
                fetchRecipes

              else if viewType == DeletedView then
                fetchDeleted

              else
                Cmd.none
            )

        GotTime posix ->
            ( { model | verseIndex = verseIndexFromPosix posix, season = seasonFromMonth (Time.toMonth Time.utc posix) }, Cmd.none )

        UpdateFilter val ->
            ( { model | filterText = val }, Cmd.none )

        SetActiveTag tag ->
            ( { model | activeTag = tag }, Cmd.none )

        UpdateScrapeUrl val ->
            ( { model | scrapeUrl = val }, Cmd.none )

        RunScrape ->
            ( { model | status = "Scraping recipe data..." }, scrapeRecipe model.scrapeUrl )

        ScrapeResult res ->
            case res of
                Ok r ->
                    let
                        empty =
                            String.trim r.title == "" && List.isEmpty r.ingredients && List.isEmpty r.instructions
                    in
                    ( { model
                        | title = r.title
                        , ingredients = String.join "\n" r.ingredients
                        , instructions = String.join "\n" r.instructions
                        , notes = r.notes
                        , status =
                            if empty then
                                "Scrape found no recipe data. Try a different URL, or enter the recipe manually."

                            else
                                "Scraped! Review and edit the recipe, then Save."
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

        FetchDeleted ->
            ( model, fetchDeleted )

        DeletedFetched res ->
            case res of
                Ok list ->
                    ( { model | deleted = list }, Cmd.none )

                Err _ ->
                    ( model, Cmd.none )

        ToggleDeletedSelect title ->
            ( { model | selectedDeleted = toggleSelected title model.selectedDeleted }, Cmd.none )

        ToggleSelectAll ->
            let
                allTitles =
                    List.map .title model.deleted

                allSelected =
                    not (List.isEmpty allTitles) && List.length model.selectedDeleted == List.length allTitles
            in
            ( { model | selectedDeleted = if allSelected then [] else allTitles }, Cmd.none )

        RestoreSelected ->
            ( { model | status = "Restoring..." }, Cmd.batch (List.map restoreRequest model.selectedDeleted) )

        Restored _ ->
            ( { model | selectedDeleted = [], status = "Restored." }, fetchDeleted )

        StartPurge intent ->
            ( { model | purgeIntent = Just intent, purgeConfirm = "" }, Cmd.none )

        UpdatePurgeConfirm val ->
            ( { model | purgeConfirm = val }, Cmd.none )

        CancelPurge ->
            ( { model | purgeIntent = Nothing, purgeConfirm = "" }, Cmd.none )

        ConfirmPurge ->
            case model.purgeIntent of
                Just PurgeSelected ->
                    ( { model | purgeIntent = Nothing, purgeConfirm = "", status = "Purging..." }, Cmd.batch (List.map purgeRequest model.selectedDeleted) )

                Just PurgeAll ->
                    ( { model | purgeIntent = Nothing, purgeConfirm = "", status = "Purging..." }, purgeAllRequest )

                Nothing ->
                    ( model, Cmd.none )

        Purged _ ->
            ( { model | selectedDeleted = [], status = "Purged." }, fetchDeleted )

        ScaleRecipe factor ->
            case model.currentView of
                ReadingView recipe ->
                    ( { model | scaleFactor = factor }, scaleRequest recipe.title factor )

                _ ->
                    ( model, Cmd.none )

        Scaled res ->
            case res of
                Ok r ->
                    ( { model | scaled = Just r }, Cmd.none )

                Err _ ->
                    ( { model | status = "Failed to scale." }, Cmd.none )

        ResetScale ->
            ( { model | scaled = Nothing, scaleFactor = "1" }, Cmd.none )



-- 3. VIEW


view : Model -> Html Msg
view model =
    div [ class "page", attribute "data-theme" (seasonToString model.season) ]
        [ -- APP BAR
          div [ class "appbar" ]
            [ div [ class "appbar-brand" ]
                [ span [ class "appbar-title" ] [ text "Recipe Tracker" ]
                , span [ class "appbar-subtitle" ] [ text "Morris Family" ]
                ]
            , div [ class "appbar-nav" ]
                [ button (onClick (SetView EntryView) :: navStyle) [ text "Add" ]
                , button (onClick (SetView ListView) :: navStyle) [ text "Recipe Box" ]
                , button (onClick (SetView DeletedView) :: navStyle) [ text "Deleted" ]
                , button (onClick (SetView FeedbackView) :: navStyle) [ text "Feedback" ]
                ]
            ]
        , div [ class "container" ]
            [ case model.currentView of
                EntryView ->
                    viewEntryForm model

                ListView ->
                    viewRecipeList model

                ReadingView recipe ->
                    viewReader model recipe

                DeletedView ->
                    viewDeleted model

                FeedbackView ->
                    viewFeedback
            , p [ class "status" ] [ text model.status ]
            ]
        ]


viewEntryForm : Model -> Html Msg
viewEntryForm model =
    div []
        [ -- SCRAPER SECTION
          div [ class "scraper" ]
            [ h4 [ class "scraper-title" ] [ text "✨ Import from Web" ]
            , div [ class "scraper-row" ]
                [ input [ placeholder "Paste recipe URL here...", value model.scrapeUrl, onInput UpdateScrapeUrl, class "scraper-input" ] []
                , button [ onClick RunScrape, class "scraper-btn" ] [ text "Import" ]
                ]
            ]
        , -- MANUAL ENTRY SECTION
          div [ class "card" ]
            [ div [ class "card-head" ]
                [ h2 [ class "card-title" ]
                    [ text
                        (if String.startsWith "Editing" model.status then
                            "Edit Recipe"

                         else
                            "Morris Recipe Entry"
                        )
                    ]
                , button [ onClick ClearForm, class "clear-btn" ] [ text "Clear All" ]
                ]
            , input (placeholder "Title" :: value model.title :: onInput UpdateTitle :: inputStyle) []
            , input (placeholder "Tags (e.g. Dinner, Beef, Slow Cooker)" :: value model.tags :: onInput UpdateTags :: inputStyle) []
            , textarea (placeholder "Ingredients (One per line)" :: rows 6 :: value model.ingredients :: onInput UpdateIngredients :: inputStyle) []
            , textarea (placeholder "Instructions (One per line)" :: rows 6 :: value model.instructions :: onInput UpdateInstructions :: inputStyle) []
            , input (placeholder "Notes" :: value model.notes :: onInput UpdateNotes :: inputStyle) []
            , div [ class "form-row" ]
                [ button [ onClick SaveRecipe, class "save-btn" ] [ text "Save Recipe" ]
                , if String.startsWith "Editing" model.status then
                    button [ onClick CancelEdit, class "cancel-btn" ] [ text "Cancel" ]

                  else
                    text ""
                ]
            ]
        ]


viewFeedback : Html Msg
viewFeedback =
    div [ class "card" ]
        [ div [ class "card-head" ]
            [ h2 [ class "card-title" ] [ text "Send Feedback" ]
            , button [ onClick (SetView ListView), class "clear-btn" ] [ text "Cancel" ]
            ]
        , p [ class "muted" ] [ text "Questions, bugs, or ideas — we read every message." ]
        , Html.form
            [ action "https://submit-form.com/SftusnkmF"
            , method "post"
            , target "_blank"
            ]
            [ label [ for "name", class "feedback-label" ] [ text "Name" ]
            , input [ type_ "text", id "name", name "name", placeholder "Name", required True, class "input" ] []
            , label [ for "email", class "feedback-label" ] [ text "Email" ]
            , input [ type_ "email", id "email", name "email", placeholder "Email", required True, class "input" ] []
            , label [ for "message", class "feedback-label" ] [ text "Message" ]
            , textarea [ id "message", name "message", placeholder "Message", required True, rows 6, class "input" ] []
            , button [ type_ "submit", class "done-btn" ] [ text "Send" ]
            ]
        , p [ class "muted" ] [ text "Submitting opens a confirmation in a new tab." ]
        ]


viewRecipeList : Model -> Html Msg
viewRecipeList model =
    let
        searchLower =
            String.toLower model.filterText

        matchesSearch r =
            String.contains searchLower (String.toLower r.title)
                || List.any (\ing -> String.contains searchLower (String.toLower ing)) r.ingredients
                || List.any (\tag -> String.contains searchLower (String.toLower tag)) r.tags
                || String.contains searchLower (String.toLower r.notes)

        matchesTag r =
            model.activeTag == "" || List.member model.activeTag r.tags

        filteredRecipes =
            List.filter (\r -> matchesTag r && matchesSearch r) model.recipes

        allTags =
            model.recipes
                |> List.concatMap .tags
                |> List.map String.trim
                |> List.filter (not << String.isEmpty)
                |> Set.fromList
                |> Set.toList

        tagCount tag =
            List.length (List.filter (\r -> List.member tag r.tags) model.recipes)

        totalLabel =
            let
                n =
                    List.length model.recipes
            in
            if n == 1 then
                "1 recipe"

            else
                String.fromInt n ++ " recipes"

        verse =
            getDailyVerse model.verseIndex
    in
    div []
        [ h2 [ class "list-title" ] [ text "Morris Family Recipe Box" ]
        , p [ class "muted" ] [ text totalLabel ]
        , -- SCRIPTURE CARD
          div [ class "scripture" ]
            [ p [ class "scripture-text" ] [ text ("\"" ++ verse.text ++ "\"") ]
            , p [ class "scripture-ref" ] [ text verse.ref ]
            ]
        , input (placeholder "Search recipes..." :: value model.filterText :: onInput UpdateFilter :: searchStyle) []
        , -- CATEGORY FILTER CHIPS
          div [ class "chip-row" ]
            (tagChip "All" (model.activeTag == "") (SetActiveTag "")
                :: List.map (\t -> tagChip (t ++ " (" ++ String.fromInt (tagCount t) ++ ")") (model.activeTag == t) (SetActiveTag t)) allTags
            )
        , -- MASTER COOKBOOK BUTTONS
          div [ class "cookbook-row" ]
            [ a (href "/api/export/cookbook?booklet=false" :: target "_blank" :: masterBtnStyle "#6B705C") [ text "Letter Cookbook" ]
            , a (href "/api/export/cookbook?booklet=true" :: target "_blank" :: masterBtnStyle "#A5A58D") [ text "Booklet Cookbook" ]
            ]
        , if List.isEmpty filteredRecipes then
            div [ class "empty" ] [ text "No recipes found." ]

          else
            div [] (List.map (viewRecipeCard model.deletingTitle) filteredRecipes)
        ]


viewRecipeCard : Maybe String -> Recipe -> Html Msg
viewRecipeCard deletingTitle recipe =
    let
        isDeleting =
            deletingTitle == Just recipe.title
    in
    div [ class "recipe-card" ]
        [ strong [ class "recipe-title", onClick (OpenReader recipe) ] [ text recipe.title ]
        , p [ class "recipe-tags" ] [ text (String.join ",  " recipe.tags) ]
        , if isDeleting then
            div [ class "confirm-bar" ]
                [ span [ class "confirm-text" ] [ text "Delete?" ]
                , div [ class "btn-row" ]
                    [ button [ onClick (ExecuteDelete recipe.title), class "yes-btn" ] [ text "Yes" ]
                    , button [ onClick CancelDelete, class "no-btn" ] [ text "No" ]
                    ]
                ]

          else
            div [ class "action-row" ]
                [ a (href ("/api/export/pdf?title=" ++ recipe.title ++ "&booklet=false") :: target "_blank" :: actionBtnStyle "#6B705C") [ text "Letter" ]
                , a (href ("/api/export/pdf?title=" ++ recipe.title ++ "&booklet=true") :: target "_blank" :: actionBtnStyle "#A5A58D") [ text "Booklet" ]
                , button (onClick (EditRecipe recipe) :: actionBtnStyle "#d4a373") [ text "Edit" ]
                , button (onClick (ConfirmDelete recipe.title) :: [ class "btn-danger-ghost" ]) [ text "Delete" ]
                ]
        ]


viewReader : Model -> Recipe -> Html Msg
viewReader model recipe =
    let
        display =
            Maybe.withDefault recipe model.scaled

        scaled =
            model.scaled /= Nothing
    in
    div [ class "reader-card" ]
        [ div [ class "reader-head" ]
            [ button [ onClick (SetView ListView), class "back-btn" ] [ text "← Back" ]
            , button [ onClick (EditRecipe recipe), class "reader-edit-btn" ] [ text "Edit" ]
            ]
        , h1 [ class "reader-title" ] [ text display.title ]
        , div [ class "reader-tags" ] [ text (String.join ",  " display.tags) ]
        , div [ class "scale-row" ]
            [ scaleButton model "½×" "0.5"
            , scaleButton model "1×" "1"
            , scaleButton model "2×" "2"
            , scaleButton model "3×" "3"
            ]
        , if scaled then
            div [ class "scale-banner" ]
                [ span [] [ text ("⚠ ×" ++ model.scaleFactor ++ " — temporarily scaled") ]
                , button [ onClick ResetScale, class "scale-reset" ] [ text "Reset" ]
                ]

          else
            text ""
        , h3 [ class "section-heading" ] [ text "Ingredients" ]
        , ul [ class "ingredient-list" ] (List.map (\ing -> li [] [ text ing ]) display.ingredients)
        , h3 [ class "section-heading-spaced" ] [ text "Instructions" ]
        , ol [ class "instruction-list" ] (List.map (\inst -> li [ class "instruction-item" ] [ text (cleanInstruction inst) ]) display.instructions)
        , if String.isEmpty display.notes then
            text ""

          else
            div [ class "notes-box" ]
                [ h4 [ class "notes-heading" ] [ text "Notes" ], renderNotes display.notes ]
        , button [ onClick (SetView ListView), class "done-btn" ] [ text "Done Reading" ]
        ]



viewDeleted : Model -> Html Msg
viewDeleted model =
    let
        selectedCount =
            List.length model.selectedDeleted
    in
    div []
        [ h2 [ class "list-title" ] [ text "Deleted Recipes" ]
        , p [ class "muted" ] [ text "Restore to bring a recipe back, or purge to remove it permanently." ]
        , if List.isEmpty model.deleted then
            div [ class "empty" ] [ text "No deleted recipes." ]

          else
            div []
                [ div [ class "toolbar-row" ]
                    [ button (onClick ToggleSelectAll :: [ class "btn-gray" ]) [ text "Select All" ]
                    , button (disabled (selectedCount == 0) :: onClick RestoreSelected :: actionBtnStyle "#6B705C") [ text ("Restore (" ++ String.fromInt selectedCount ++ ")") ]
                    , button (disabled (selectedCount == 0) :: onClick (StartPurge PurgeSelected) :: actionBtnStyle "#a00") [ text "Delete Forever" ]
                    , button (onClick (StartPurge PurgeAll) :: actionBtnStyle "#a00") [ text "Purge All" ]
                    ]
                , div [] (List.map (viewDeletedCard model.selectedDeleted) model.deleted)
                ]
        , viewPurgeConfirm model
        ]


viewDeletedCard : List String -> Recipe -> Html Msg
viewDeletedCard selected recipe =
    div [ class "deleted-card" ]
        [ label [ class "deleted-label" ]
            [ input [ type_ "checkbox", checked (List.member recipe.title selected), onCheck (\_ -> ToggleDeletedSelect recipe.title) ] []
            , div []
                [ strong [ class "deleted-title" ] [ text recipe.title ]
                , p [ class "deleted-tags" ] [ text (String.join ",  " recipe.tags) ]
                ]
            ]
        ]


viewPurgeConfirm : Model -> Html Msg
viewPurgeConfirm model =
    case model.purgeIntent of
        Nothing ->
            text ""

        Just intent ->
            let
                target =
                    case intent of
                        PurgeSelected ->
                            "the selected recipes"

                        PurgeAll ->
                            "ALL deleted recipes"

                confirmed =
                    model.purgeConfirm == "yesdeary"
            in
            div [ class "purge-box" ]
                [ p [ class "purge-warning" ] [ text ("This permanently deletes " ++ target ++ ". Type \"yesdeary\" to confirm.") ]
                , input (placeholder "yesdeary" :: value model.purgeConfirm :: onInput UpdatePurgeConfirm :: inputStyle) []
                , div [ class "btn-row" ]
                    [ button (disabled (not confirmed) :: onClick ConfirmPurge :: actionBtnStyle "#a00") [ text "Permanently Delete" ]
                    , button (onClick CancelPurge :: [ class "btn-gray" ]) [ text "Cancel" ]
                    ]
                ]


toggleSelected : String -> List String -> List String
toggleSelected title selected =
    if List.member title selected then
        List.filter (\t -> t /= title) selected

    else
        title :: selected



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


getDailyVerse : Int -> { text : String, ref : String }
getDailyVerse index =
    bibleVerses |> List.drop index |> List.head |> Maybe.withDefault { text = "", ref = "" }


verseIndexFromPosix : Time.Posix -> Int
verseIndexFromPosix posix =
    modBy (List.length bibleVerses) (Time.posixToMillis posix // 86400000)


seasonFromMonth : Time.Month -> Season
seasonFromMonth month =
    case month of
        Time.Mar ->
            Spring

        Time.Apr ->
            Spring

        Time.May ->
            Spring

        Time.Jun ->
            Summer

        Time.Jul ->
            Summer

        Time.Aug ->
            Summer

        Time.Sep ->
            Autumn

        Time.Oct ->
            Autumn

        Time.Nov ->
            Autumn

        _ ->
            Winter


seasonToString : Season -> String
seasonToString season =
    case season of
        Spring ->
            "spring"

        Summer ->
            "summer"

        Autumn ->
            "autumn"

        Winter ->
            "winter"


cleanInstruction : String -> String
cleanInstruction inst =
    let
        re =
            Regex.fromString "^\\d+[\\.\\)]\\s*" |> Maybe.withDefault Regex.never
    in
    Regex.replace re (\_ -> "") (String.trim inst)


urlRegex : Regex.Regex
urlRegex =
    Maybe.withDefault Regex.never (Regex.fromString "https?://[^\\s]+")


linkify : String -> List (Html Msg)
linkify str =
    linkifyFrom str 0 (Regex.find urlRegex str)


linkifyFrom : String -> Int -> List Regex.Match -> List (Html Msg)
linkifyFrom str cursor matches =
    case matches of
        [] ->
            [ text (String.slice cursor (String.length str) str) ]

        m :: rest ->
            let
                url =
                    m.match

                nextCursor =
                    m.index + String.length url
            in
            [ text (String.slice cursor m.index str)
            , a [ href url, target "_blank", class "link" ] [ text url ]
            ]
                ++ linkifyFrom str nextCursor rest


renderNotes : String -> Html Msg
renderNotes notes =
    div []
        (notes
            |> String.split "\n"
            |> List.filter (\line -> String.trim line /= "")
            |> List.map (\line -> p [ class "note-line" ] (linkify line))
        )



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


fetchDeleted : Cmd Msg
fetchDeleted =
    Http.get { url = "/api/deleted", expect = Http.expectJson DeletedFetched (Decode.list recipeDecoder) }


scaleRequest : String -> String -> Cmd Msg
scaleRequest title factor =
    Http.get { url = "/api/scale?title=" ++ Url.percentEncode title ++ "&factor=" ++ factor, expect = Http.expectJson Scaled recipeDecoder }


restoreRequest : String -> Cmd Msg
restoreRequest title =
    Http.post { url = "/api/restore?title=" ++ Url.percentEncode title, body = Http.emptyBody, expect = Http.expectWhatever Restored }


purgeRequest : String -> Cmd Msg
purgeRequest title =
    Http.post { url = "/api/purge?title=" ++ Url.percentEncode title ++ "&confirm=yesdeary", body = Http.emptyBody, expect = Http.expectWhatever Purged }


purgeAllRequest : Cmd Msg
purgeAllRequest =
    Http.post { url = "/api/purge-all?confirm=yesdeary", body = Http.emptyBody, expect = Http.expectWhatever Purged }



-- STYLES


inputStyle =
    [ class "input" ]


searchStyle =
    [ class "search" ]


actionBtnClass : String -> String
actionBtnClass c =
    case c of
        "#A5A58D" ->
            "btn-muted"

        "#d4a373" ->
            "btn-tan"

        "#a00" ->
            "btn-red"

        _ ->
            "btn-sage"


actionBtnStyle : String -> List (Attribute Msg)
actionBtnStyle c =
    [ class (actionBtnClass c) ]


masterBtnClass : String -> String
masterBtnClass c =
    case c of
        "#A5A58D" ->
            "master-muted"

        _ ->
            "master-sage"


masterBtnStyle : String -> List (Attribute Msg)
masterBtnStyle c =
    [ class (masterBtnClass c) ]


navStyle =
    [ class "nav-btn" ]


tagChip : String -> Bool -> Msg -> Html Msg
tagChip label selected msg =
    button [ onClick msg, class ("chip " ++ (if selected then "chip-selected" else "chip-unselected")) ] [ text label ]



scaleButton : Model -> String -> String -> Html Msg
scaleButton model label factor =
    button (onClick (ScaleRecipe factor) :: actionBtnStyle (if model.scaleFactor == factor then "#6B705C" else "#A5A58D")) [ text label ]



-- 6. MAIN


main : Program () Model Msg
main =
    Browser.element
        { init = \() -> ( initialModel, Task.perform GotTime Time.now )
        , view = view
        , update = update
        , subscriptions = \_ -> Time.every 60000 GotTime
        }

