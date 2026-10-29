// Boton "Imprimir o guardar como PDF" del reporte (spec US-034). El boton se dibuja oculto: sin
// JavaScript no aparece un boton que no hace nada (RN-034-3). Este script lo muestra y, al tocarlo,
// abre el dialogo de impresion del navegador y nada mas (RN-034-1).
document.addEventListener("DOMContentLoaded", function () {
  var boton = document.getElementById("imprimir");
  if (!boton) {
    return;
  }
  boton.hidden = false;
  boton.addEventListener("click", function () {
    window.print();
  });
});
