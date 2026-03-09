function OrderPage({ orders, currentUser, loading }) {
  return (
    <section className="page">
      <div className="page-head">
        <div>
          <p className="tag">Queue Tracking</p>
          <h2>Order Status</h2>
          <p>
            {currentUser
              ? `Showing orders for ${currentUser.name}`
              : "Login to view your order queue status."}
          </p>
        </div>
      </div>

      {loading ? <p className="empty">Loading orders...</p> : null}
      {!loading && currentUser && !orders.length ? (
        <p className="empty">No orders yet. Checkout from cart to create one.</p>
      ) : null}

      <div className="stack">
        {orders.map((order) => (
          <article key={order.id} className="order-card">
            <div>
              <h4>Order #{order.id}</h4>
              <p>
                {order.menuName} • Qty {order.qty}
              </p>
              <p>Created by user #{order.userId}</p>
            </div>
            <span className={`status-chip status-${String(order.status || "CREATED").replace(/\s+/g, "-").toLowerCase()}`}>
              {order.status || "CREATED"}
            </span>
          </article>
        ))}
      </div>
    </section>
  )
}

export default OrderPage
