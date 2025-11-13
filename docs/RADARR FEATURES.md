# INTEGRACIÓN RADARR

## Página de configuración y funcionalidad

La página de configuración de radar tendrá varias secciones de configuración

- [ ] botones:
  - [ ] Toggle para activar desactivar radarr, desactivado desactiva todos los campos y elimina radarr, si se activa levanta radarr
  - [ ] botón Aplicar (apply) (sustituye guardar configuración): aplica los cambios, y reconstruye el contenedor si corresponde
  - [ ] botón start/stop: aparece uno de los dos dinámicamente según el estado del contenedor
- [ ] Configuraciones:
  - [ ] configuración de docker con los campos (cualquiera de estas opciones eliminará el contenedor, guardará la configuración y levantará de nuevo radarr con la nueva configuración) :
    - [ ] tag de la imagen docker (se podrán cargar dinámicamente los últimos tags para seleccionarlos) si está vacío usará latest
    - [ ] puerto expuesto: si está vacío no se expone, si está relleno se expone el puerto expuesto el puerto interno será del de radarr 7878
    - [ ] restart policy: desplegable con las opciones
    - [ ] config path: ruta a la configuración en el host, por defecto estará relleno y el valor será ./volumes/radarr/config, esta carpeta debe estar en la raiz de este repositorio
  - [ ] Configuración de radarr (estas opciones configurarán directamente radarr !TODO)
    - [ ] !TODO
  - [ ] Proxy/Acceso externo: configuración de dominio:
    - [ ] subdominio: si está vacío usará radarr, por defecto vacío
    - [ ] dominio: si está vacío usará la configuración global, por defecto vacío

------
a tener en cuenta
-por defecto el valor del puerto debe estar vacío (puerto sin exponer)
-el host port debe poder quedarse vacío, no a 0, vacío
-los botones de stop start se intercambian según si el contenedor está en ejecución o no
-el botón de restart se lanzará con --forcde-recreate, haciendo un kill y un down antes
-docker image no se podrá cambiar, pero el tag sí tendrá un desplegable con las versiones disponibles además de latest como primera opción y por defecto
-añade otro botón reset que borre la configuración (volumen config) y recree el contenedor con los valores por defecto
-si cambia cualquier configuración de docker al guardar se debe recrear el contenedor, con pull si cambia el tag, y sin pull si no cambia el tag,
