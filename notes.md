# Heirarchy
MainModel
|-Shows All-> Status (Shows current status) (Not a model)
|- - - - - -> ContentModel (Swaps between various screens)
|             |-Swaps Between-> LibraryModel
|             |- - - - - - - -> AlbumModel
|			  +- - - - - - - -> QueueModel
+- - - - - -> ControllerModel (Controls Music)
