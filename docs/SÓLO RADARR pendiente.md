SÓLO RADARR

borra de aquí lo que esté implementado

------
por defecto el valor del puerto debe estar vacío (puerto sin exponer)
el host port debe poder quedarse vacío, no a 0, vacío
los botones de stop start se reemplazan según si el contenedor está en ejecución o no
el botón de restart se lanzará con recreate
docker image no se podrá cambiar, pero el tag sí tendrá un desplegable con las versiones disponibles además de latest como primera opción y por defecto
añade otro botón reset que borre la configuración (volumen config) y recree el contenedor
si cambia cualquier configuración de docker al guardar se debe recrear el contenedor con pull si cambia el tag, y sin pull si no cambia el tag, 
los volúmenes de mediacheky pueden ser uno solo o hacen falta dos?