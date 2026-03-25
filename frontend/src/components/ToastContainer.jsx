function ToastContainer({ toasts, onClose }) {
  return (
    <div className="toast-stack" aria-live="polite" aria-atomic="true">
      {toasts.map((toast) => (
        <article key={toast.id} className={`toast toast-${toast.type}`}>
          <div>
            <strong>{toast.title}</strong>
            {toast.message ? <p>{toast.message}</p> : null}
          </div>
          <button className="ghost" onClick={() => onClose(toast.id)} aria-label="Close notification">
            x
          </button>
        </article>
      ))}
    </div>
  )
}

export default ToastContainer
