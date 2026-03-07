import { useState } from "react"
import AddMenuForm from "../components/AddMenuForm"

const NEXT_STATUS = {
  CREATED: "PREPARING",
  PREPARING: "READY",
  READY: "PICKED_UP",
}

function AdminMenuPage({
  menus,
  onCreate,
  onUpdate,
  onDelete,
  onRefresh,
  orders,
  ordersLoading,
  onRefreshOrders,
  onAdvanceOrderStatus,
}) {
  const [editingId, setEditingId] = useState(null)
  const [message, setMessage] = useState("")
  const [orderMessage, setOrderMessage] = useState("")

  const handleCreate = async (payload) => {
    try {
      await onCreate(payload)
      setMessage("Menu added successfully.")
    } catch (error) {
      setMessage(error.message)
    }
  }

  const handleUpdate = async (id, payload) => {
    try {
      await onUpdate(id, payload)
      setEditingId(null)
      setMessage("Menu updated.")
    } catch (error) {
      setMessage(error.message)
    }
  }

  const handleDelete = async (id) => {
    try {
      await onDelete(id)
      setMessage("Menu deleted.")
    } catch (error) {
      setMessage(error.message)
    }
  }

  const handleAdvanceStatus = async (order) => {
    const next = NEXT_STATUS[order.status]
    if (!next) return

    try {
      await onAdvanceOrderStatus(order.id, next)
      setOrderMessage(`Order #${order.id} moved to ${next}.`)
    } catch (error) {
      setOrderMessage(error.message)
    }
  }

  return (
    <section className="page">
      <div className="page-head">
        <div>
          <p className="tag">Vendor Admin</p>
          <h2>Menu Management</h2>
          <p>Maintain item list and pricing for the single canteen store.</p>
        </div>
        <button className="ghost" onClick={onRefresh}>Reload</button>
      </div>

      <AddMenuForm submitLabel="Add Menu" onSubmit={handleCreate} />
      {message ? <p className="hint-text">{message}</p> : null}

      <div className="stack">
        {menus.map((item) => (
          <article key={item.id} className="admin-row">
            {editingId === item.id ? (
              <AddMenuForm
                initialValues={{ name: item.name, price: item.price, image_url: item.image_url }}
                submitLabel="Save Changes"
                onSubmit={(payload) => handleUpdate(item.id, payload)}
                onCancel={() => setEditingId(null)}
              />
            ) : (
              <>
                <div>
                  <h4>{item.name}</h4>
                  <p>THB {Number(item.price).toFixed(2)}</p>
                  <p className="img-url">{item.image_url ? item.image_url : "No image URL"}</p>
                </div>
                <div className="row-actions">
                  <button className="ghost" onClick={() => setEditingId(item.id)}>Edit</button>
                  <button className="danger" onClick={() => handleDelete(item.id)}>Delete</button>
                </div>
              </>
            )}
          </article>
        ))}
      </div>

      <div className="page-head" style={{ marginTop: 22 }}>
        <div>
          <p className="tag">Order Queue</p>
          <h2>Order Lifecycle</h2>
          <p>Advance order statuses from CREATED to PICKED_UP.</p>
        </div>
        <button className="ghost" onClick={onRefreshOrders}>Reload Orders</button>
      </div>

      {orderMessage ? <p className="hint-text">{orderMessage}</p> : null}
      {ordersLoading ? <p className="empty">Loading orders...</p> : null}
      {!ordersLoading && !orders.length ? <p className="empty">No orders yet.</p> : null}

      <div className="stack">
        {orders.map((order) => {
          const next = NEXT_STATUS[order.status]
          return (
            <article key={order.id} className="order-card">
              <div>
                <h4>Order #{order.id}</h4>
                <p>
                  {order.menuName} • Qty {order.qty}
                </p>
                <p>User #{order.userId}</p>
              </div>

              <div className="row-actions">
                <span className={`status-chip status-${String(order.status || "CREATED").replace(/\s+/g, "-").toLowerCase()}`}>
                  {order.status || "CREATED"}
                </span>
                {next ? <button className="ghost" onClick={() => handleAdvanceStatus(order)}>Mark {next}</button> : null}
              </div>
            </article>
          )
        })}
      </div>
    </section>
  )
}

export default AdminMenuPage
