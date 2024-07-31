from matplotlib.pyplot import savefig, show, subplots
from numpy import arange

def main():
    # Data
    libraries = ['Sarama', 'Kafka Go', 'Franz Go']
    # message_sizes = ['1000*1KB', '1000*5KB', '1000*10KB']
    # sarama_avg = [0.001590931, 0.001581500, 0.00179425]
    # kafka_go_avg = [0.004126457,0.002186847,0.00203725]
    # franz_go_avg = [0.001004125,0.000790292,0.001044556]

    # message_sizes = ['1w*1KB', '1w*5KB', '1w*10KB']
    # sarama_avg = [0.018665097, 0.020472278, 0.020991638]
    # kafka_go_avg = [0.018903389,0.019190222,0.027173360]
    # franz_go_avg = [0.006428249,0.008191361,0.006253777]


    # message_sizes = ['10w*1KB', '10w*5KB', '10w*10KB']
    # sarama_avg = [0.093369583, 0.099443333, 0.096334222]
    # kafka_go_avg = [0.1724541803,0.21072552,0.237053222]
    # franz_go_avg = [0.279305236,0.254314583,0.234592708]

    message_sizes = ['100w*1KB(1)', '100w*1KB(2)', '100w*1KB(3)']
    sarama_avg = [2.179512125, 0.441739958, 0.73185]
    kafka_go_avg = [5.8538619579999995,1.159171459,2.317493083]
    franz_go_avg = [2.309908208,3.172073625,4.154107125]

    x = arange(3)  # the label locations
    width = 0.2  # the width of the bars

    fig, ax = subplots(figsize=(10, 6))

    rects1 = ax.bar(x - width, sarama_avg, width, label='Sarama')
    rects2 = ax.bar(x, kafka_go_avg, width, label='Kafka Go')
    rects3 = ax.bar(x + width, franz_go_avg, width, label='Franz Go')

    # Add some text for labels, title and custom x-axis tick labels, etc.
    ax.set_xlabel('Message Sizes')
    ax.set_ylabel('Average Time (seconds)')
    ax.set_title('Average Time to Produce Messages by Kafka Clients 100w')
    ax.set_xticks(x)
    ax.set_xticklabels(message_sizes)
    ax.legend()

    fig.tight_layout()

    # Save plot to a file
    savefig('/Users/xuhaidong/Downloads/kafka_clients_benchmark_100w.png')

    # Display plot
    show()


if __name__ == "__main__":
    main()