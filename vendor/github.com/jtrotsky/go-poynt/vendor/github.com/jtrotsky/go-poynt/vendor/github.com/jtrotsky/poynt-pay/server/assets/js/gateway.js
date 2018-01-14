// Handles basic communication with the Vend postMessage Payments API.
// Documentation: https://developers.vendhq.com/documentation/client/payment.html
//
//

// On first page-load request additional sale data from Vend.
window.addEventListener("load", onPageLoad, false);

// On initial load of modal, first configure window appearance, then request
// further sale data from Vend to start the payment process.
 function onPageLoad() {
   setupStep();
   dataStep();
 }

// Listen for postMessage events from Vend, this is how each step of the payment
// process is communicated.
window.addEventListener("message", function(event) {
  // Read event data from Vend.
  const data = JSON.parse(event.data);
  // Store payment amount.
  if (data.payment.amount) {
    amount = data.payment.amount;
  }
  // Store UUID from the register that sent the payment.
  if (data.payment.register_id) {
    regiserID = data.payment.register_id;
  }

  // If we get anything back from Vend other than the DATA step, something has
  // gone wrong.
  if (data.step != 'DATA') {
    // TODO: Error handling.
    alert("Strange response from Vend.");
  } else {
      // Instruct cashier to wait for customer input as payment is being sent to
      // terminal.
      $('#statusTextContainer').append("Tap or Insert Card");

      // Request /pay endpoint to send amount to terminal and wait for respnse.
      $.ajax({
        type: "GET",
        url: "pay",
        data: { "amount": amount, "origin": getQueryString()['origin'] },
      })
      // If AJAX call is completed, then
      .done(function(response) {
        // TODO: check if response is not empty before parsing.
        responseBody = JSON.parse(response);
        // Always log repsonse body.
        console.log(responseBody);
        // Make sure status text is cleared.
        $('#statusTextContainer').empty();
        // Read transaction status and act appropriately.
        checkTerminalResponse(responseBody);
      })
      // Likeliest reason for this will be communication. If the network is down
      // or the like.
      // TODO: Add fallback function.
      .fail(function(error) {
        // Always log error body.
        console.log(error);
        // Make sure status text is cleared.
        $('#statusTextContainer').empty();
        $('#statusTextContainer').append("Transaction Failed")
        // Quit window, giving cashier chance to try again.
        // TODO: Display retry button or transaction status check button.
        window.setTimeout(exitStep, 2000)
      })
      .always(function() {
        // TODO: Remove
        // Nothing to do always.
      })
  }}, false);

// Check response from POYNT terminal.
function checkTerminalResponse(responseBody) {
  // responseBody is the terminal response JSON, containing the request
  // reference and transaction status, example:
  // {
	//   "referenceId": "532666cd-a992-475c-a004-15f0e804345f",
	//   "status": "CANCELED"
  // }
  // Check response status field.
  switch (responseBody.status) {
    case 'AUTHORIZED':
    // Fallthrough
    // TODO: AUTH step?
    case 'CANCELED':
      $('#statusTextContainer').append("Transaction Cancelled")
      window.setTimeout(exitStep, 2500)
      break;
    case 'COMPLETED':
      $('#statusTextContainer').append("Transaction Accepted")
      window.setTimeout(acceptStep, 2500)
      break;
    case 'FAILED':
      $('#statusTextContainer').append("Transaction Failed")
      window.setTimeout(exitStep, 2500)
      // Fallthrough
    case 'REFUNDED':
      // Fallthrough
      // TODO: Have to build refund handling.
    case 'VOIDED':
      // Fallthrough
      // TODO: What is being voided?
    default:
      // Don't know what we got, or something went wrong, so log it.
      console.log(responseBody);
      break;
  }
};

// Send payload to the Payments API.
 function sendObjectToVend(object) {
   var receiver = window.opener !== null ? window.opener : window.parent;
   receiver.postMessage(JSON.stringify(object), "*");
 };

 // ACCEPT: Trigger a successful transaction. If the payment type supports
 // printing (and it’s enabled) an approved transaction receipt will also print.
 function acceptStep() {
   sendObjectToVend({
     step: "ACCEPT",
     success: true,
   });
 };

// DATA: Request additional information from Vend about the sale and payment.
function dataStep() {
  sendObjectToVend({
    step: "DATA",
    success: true,
    // TODO: What is "name" key even for?
    name: "payment"
  });
};

// EXIT: Cleanly exit the process. Doesn’t close the window but closes all
// other dialogs including the payment modal/iFrame and unbinds postMessage
// handling.
 function exitStep() {
   sendObjectToVend({
     step: "EXIT",
     success: true
   });
 };

 // SETUP: Customize the payment dialog.
 // At this stage removing close button to prevent retailers
 // from prematurely closing the modal, and thus interrupting the payment flow
 // without a clean exit.
  function setupStep() {
   sendObjectToVend({
     step: "SETUP",
     success: true,
     setup:
     {
       enable_close: false
     }
   });
 };

 // Get query parameters from the URL. Vend passes "amount" and "origin".
 function getQueryString() {
   var result = {}, queryString = location.search.slice(1),
   re = /([^&=]+)=([^&]*)/g, m;

   while (m = re.exec(queryString)) {
     result[decodeURIComponent(m[1])] = decodeURIComponent(m[2]);
   }
   return result;
 }
