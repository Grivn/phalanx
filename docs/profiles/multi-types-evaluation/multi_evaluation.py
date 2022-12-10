import matplotlib.pyplot as plt

bsize = [
    ('Phalanx-HS', [
        (4, 116.31),
        (6, 112.94),
        (8, 107.04),
        (10, 91.3),
    ], '-o', 'coral'),
    ('HS', [
        (4, 20.1294),
        (6, 17.3645),
        (8, 15.1018),
        (10, 13.7026),
    ], '-o', 'brown'),
    ('Phalanx-TCHS', [
        (4, 124.1294),
        (6, 114.3645),
        (8, 113.1018),
        (10, 95.7026),
    ], '-p', 'darkseagreen'),
    ('TCHS', [
        (4, 21.1294),
        (6, 17.3645),
        (8, 16.1018),
        (10, 14.7026),
    ], '-o', 'gold'),
    ('Phalanx-SL', [
        (4, 31.1294),
        (6, 10.3645),
        (8, 52.1018),
        (10, 84.7026),
    ], '-o', 'wheat'),
    ('SL', [
        (4, 15.1294),
        (6, 6.3645),
        (8, 8.1018),
        (10, 12.7026),
    ], '-o', 'indianred'),
    ('Phalanx-FHS', [
        (4, 124.1294),
        (6, 116.3645),
        (8, 112.1018),
        (10, 91.7026),
    ], '-o', 'royalblue'),
    ('FHS', [
        (4, 22.1294),
        (6, 17.3645),
        (8, 16.1018),
        (10, 15.7026),
    ], '-o', 'slateblue'),
    ('Phalanx-LBFT', [
        (4, 37.1294),
        (6, 11.3645),
        (8, 44.1018),
        (10, 83.7026),
    ], '-o', 'violet'),
    ('LBFT', [
        (4, 11.1294),
        (6, 6.3645),
        (8, 8.1018),
        (10, 12.7026),
    ], '-o', 'teal'),
]


def do_plot():
    f = plt.figure(1, figsize=(8, 4))
    plt.clf()
    ax = f.add_subplot(1, 1, 1)
    for name, entries, style, color in bsize:
        throughput = []
        replica_no = []
        for c, t in entries:
            throughput.append(t)
            replica_no.append(c)
        ax.plot(replica_no, throughput, style, color=color, label='%s' % name, markersize=6, alpha=0.8)
    plt.legend(loc=2, bbox_to_anchor=(1.05, 1.0), borderaxespad=0.)
    plt.grid(linestyle='--', alpha=0.3)
    plt.ylim([0, 150])
    plt.ylabel('Throughput ($10^4$tx/s)')
    plt.xlabel('Number of Consensus Nodes.')
    plt.tight_layout()
    plt.savefig('multi_evaluation.pdf', format='pdf')
    plt.show()


if __name__ == '__main__':
    do_plot()
